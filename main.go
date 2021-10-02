package main

import (
	"os"
	"text/template"

	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"gitlab.com/rainbird-ai/sdk-go"
)

const (
	kmID = "bfaab567-6494-4dd3-bbdb-0915a32da0f7"
)

var (
	answersTemplate = template.Must(template.New("answers").
		Parse(`{{range .}}{{.Subject}} {{.Relationship}} {{.Object}} ({{.Certainty}} certainty){{end}}`))
)

// askForObject asks the user a question to find out the object value
// restricts the user to a single answer
// prohibits the user from skipping the question
func askForObject(question *sdk.Question) (sdk.QAnswer, error) {
	answer := sdk.QAnswer{
		Subject:      question.Subject,
		Relationship: question.Relationship,
		CF:           "100",
	}

	items := make([]string, 0, len(question.Concepts))
	for _, v := range question.Concepts {
		if v, ok := v.Value.(string); ok {
			items = append(items, v)
		}
	}

	var err error

	// display Rainbird's question and get an answer from the user
	if question.CanAdd {
		prompt := promptui.SelectWithAdd{
			Label:    question.Prompt,
			Items:    items,
			AddLabel: "Other",
		}

		_, answer.Object, err = prompt.Run()
	} else {
		prompt := promptui.Select{
			Label: question.Prompt,
			Items: items,
		}

		_, answer.Object, err = prompt.Run()
	}

	return answer, err
}

// makeDecision completes a single decision tree
func makeDecision(client sdk.Client, subject, relationship, object string) error {
	// create session
	session, err := client.NewSession(kmID, uuid.New().String())
	if err != nil {
		panic(err)
	}

	// make initial query
	question, answers, err := session.Query(subject, relationship, object)
	if err != nil {
		return err
	}

	for answers == nil || len(*answers) == 0 {
		// ask user a question
		questionAnswer, err := askForObject(question)
		if err != nil {
			return err
		}

		// send user response to Rainbird
		question, answers, err = session.Response([]sdk.QAnswer{questionAnswer})
		if err != nil {
			return err
		}
	}

	// display Rainbird's answer
	return answersTemplate.Execute(os.Stdout, *answers)
}

func main() {
	key, ok := os.LookupEnv("RB_API_KEY")
	if !ok {
		panic("The environment variable RB_API_KEY needs to be set to a valid Rainbird API key")
	}
	client := sdk.Client{
		APIKey:         key,
		EnvironmentURL: sdk.EnvCommunity,
	}

	err := makeDecision(client, "Fred", "speaks", "")
	if err != nil {
		panic(err)
	}
}
