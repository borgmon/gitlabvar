package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

const (
	gitlabURL = "https://gitlab.com/api/v4/projects/%v/variables"
)

type GitlabVar struct {
	Key              string `json:"key" yaml:"key"`
	Value            string `json:"value" yaml:"value"`
	EnvironmentScope string `json:"environment_scope" yaml:"environment_scope"`
	VariableType     string `json:"variable_type" yaml:"variable_type"`
}

type Varlist []*GitlabVar

var (
	importPath  string
	exportPath  string
	gitlabToken string
	projectID   string
)

func main() {
	cliApp := getCLI()
	err := cliApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getCLI() *cli.App {
	app := &cli.App{
		Name:                 "gitlabvar",
		Usage:                "Export and import your CI variable from gitlab",
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "import",
				Aliases:     []string{"i"},
				Value:       ".gitlab-ci-var.yaml",
				Usage:       "import yaml file path",
				Destination: &importPath,
			},
			&cli.StringFlag{
				Name:        "export",
				Aliases:     []string{"o"},
				Value:       ".gitlab-ci-var.yaml",
				Usage:       "export yaml file path",
				Destination: &exportPath,
			},
			&cli.StringFlag{
				Name:        "token",
				Aliases:     []string{"t"},
				Usage:       "gitlab token. scope required: api. get it from here: https://gitlab.com/-/profile/personal_access_tokens",
				Destination: &gitlabToken,
			},
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "Project ID, get it from frontpage of the project",
				Destination: &projectID,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "apply",
				Aliases: []string{"a"},
				Usage:   "apply variable yaml to gitlab project",
				Action: func(c *cli.Context) error {
					err := verifyArg()
					if err != nil {
						log.Fatal(err)
					}

					err = applyYaml()
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "get variable yaml from gitlab project",
				Action: func(c *cli.Context) error {
					err := verifyArg()
					if err != nil {
						log.Fatal(err)
					}

					err = getYaml()
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			},
			{
				Name:    "env",
				Aliases: []string{"e"},
				Usage:   "export .env file",
				Action: func(c *cli.Context) error {
					err := verifyArg()
					if err != nil {
						log.Fatal(err)
					}

					err = getDotEnv()
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			},
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "get a sample yaml file",
				Action: func(c *cli.Context) error {
					err := sampleYaml()
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			},
		},
	}
	return app
}

func verifyArg() error {
	if projectID == "" {
		return errors.New("projectID is required")
	}
	if gitlabToken == "" {
		return errors.New("token is required")
	}
	return nil
}
func getDotEnv() error {
	varlist, err := getVars()
	if err != nil {
		return err
	}
	var buf string
	for _, e := range *varlist {
		if (e.EnvironmentScope == "qa" || e.EnvironmentScope == "*") && e.VariableType == "env_var" {
			buf += strings.Replace(e.Key, "K8S_SECRET_", "", 1) + "=" + e.Value + "\n"
		}
	}
	err = ioutil.WriteFile(".env", []byte(buf), 0644)
	fmt.Println(buf, []byte(buf))
	if err != nil {
		return err
	}
	fmt.Println("exported to .env")
	return nil
}

func getYaml() error {

	varlist, err := getVars()
	if err != nil {
		return err
	}
	d, err := yaml.Marshal(varlist)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(exportPath, d, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Saved to ", exportPath)
	return nil
}

func gitlabClient(url string, method string, data []byte) ([]byte, error) {

	var payload io.Reader
	if data != nil {
		payload = bytes.NewBuffer(data)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, fmt.Sprintf(url, projectID), payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", gitlabToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		buf := new(strings.Builder)
		_, err := io.Copy(buf, res.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(buf.String())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getVars() (varlist *Varlist, err error) {
	url := gitlabURL

	body, err := gitlabClient(url, "GET", nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &varlist)
	if err != nil {
		return nil, err
	}
	return
}

func createVars(v *GitlabVar) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	url := gitlabURL

	_, err = gitlabClient(url, "POST", data)
	if err != nil {
		return err
	}
	return nil
}

func updateVars(v *GitlabVar) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	url := gitlabURL + "/" + v.Key

	_, err = gitlabClient(url, "PUT", data)

	if err != nil {
		return err
	}
	return nil
}

func deleteVars(v *GitlabVar) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	url := gitlabURL + "/" + v.Key

	_, err = gitlabClient(url, "DELETE", data)

	if err != nil {
		return err
	}
	return nil
}

func applyYaml() error {

	d, err := ioutil.ReadFile(importPath)
	if err != nil {
		return err
	}
	var newVarlist *Varlist

	err = yaml.Unmarshal(d, &newVarlist)
	if err != nil {
		return err
	}

	oldVarlist, err := getVars()
	if err != nil {
		return err
	}

	updateL, createL, deleteL := fancy(newVarlist, oldVarlist)

	if len(*updateL) == 0 && len(*createL) == 0 && len(*deleteL) == 0 {
		fmt.Println("Nothing to update")
		return nil
	}
	fmt.Println("Are you sure you want to apply the following changes:")
	if len(*updateL) != 0 {
		fmt.Println("Updates:")
		printList(updateL)
	}
	if len(*createL) != 0 {
		fmt.Println("Creates:")
		printList(createL)
	}
	if len(*deleteL) != 0 {
		fmt.Println("Deletes:")
		printList(deleteL)
	}

	fmt.Println("(y)es or any other key to cancel")

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()

	switch char {
	case 'y':
		fmt.Println("Applying...")
		break
	default:
		fmt.Println("Canceled")
		return nil
	}

	for _, e := range *updateL {
		if err = updateVars(e); err != nil {
			return err
		}
	}

	for _, e := range *createL {
		if err = createVars(e); err != nil {
			return err
		}
	}

	for _, e := range *deleteL {
		if err = deleteVars(e); err != nil {
			return err
		}
	}
	fmt.Println("Done.")
	return nil

}

func sampleYaml() error {
	sample := `- key: K8S_SECRET_{YOUR_ENV_NAME}
  value: {YOUR_ENV_VALUE}
  environment_scope: '*'
  variable_type: env_var
`
	err := ioutil.WriteFile(exportPath, []byte(sample), 0644)
	if err != nil {
		return err
	}
	return nil
}

func fancy(newL *Varlist, oldL *Varlist) (*Varlist, *Varlist, *Varlist) {
	var (
		m       = make(map[string]*GitlabVar)
		updateL = &Varlist{}
		createL = &Varlist{}
		deleteL = &Varlist{}
	)

	for _, e := range *oldL {
		m[e.Key+e.EnvironmentScope] = e
	}

	for _, e := range *newL {
		key := e.Key + e.EnvironmentScope
		if m[key] != nil {
			if !deepEq(m[key], e) {
				*updateL = append(*updateL, e)
			}
			m[key] = nil
		} else {
			*createL = append(*createL, e)
		}
	}

	for _, v := range m {
		if v != nil {
			*deleteL = append(*deleteL, v)
		}
	}
	return updateL, createL, deleteL
}

func deepEq(a *GitlabVar, b *GitlabVar) bool {
	if a.EnvironmentScope == b.EnvironmentScope && a.Key == b.Key && a.Value == b.Value && a.VariableType == b.VariableType {
		return true
	}
	return false
}

func printList(l *Varlist) {
	for _, e := range *l {
		fmt.Println(e.Key, e.EnvironmentScope, "->", e.Value)
	}
}
