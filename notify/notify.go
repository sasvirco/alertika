// Package notify contains logic for notifications
// Currently supported notifications are only via SNS topics
// Keeps track of sent notifications
// Sends OK when is no longer alerting
package notify

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	"alertika/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/rs/zerolog/log"
)

//Notification keep track of alerting rules

//ProcessNotifications - notifies subscriber of rule evaluation was succesfful
func ProcessNotifications(notifications chan *model.NotificationItem, tmpl string) error {
	var wg sync.WaitGroup
	log.Debug().Msgf("processing [%d] channel notifications", len(notifications))
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	}))
	sns := sns.New(sess)

	for n := range notifications {
		log.Debug().Msgf("notification worker will process [%s]", n.Rule.Name)
		switch n.State {
		case "Alert":
			wg.Add(1)
			go SendAlert(&wg, n, tmpl, sns)
		case "Ok":
			wg.Add(1)
			go SendOk(&wg, n, tmpl, sns)
		}

	}
	wg.Wait()
	return nil
}

//SendAlert - alerting
func SendAlert(wg *sync.WaitGroup, n *model.NotificationItem, t string, client *sns.SNS) error {
	defer wg.Done()
	log.Info().Msgf("sending Alert for [%s]", n.Rule.Name)
	msg, err := RenderMessageTemplate(n, t)
	if err != nil {
		log.Error().Msgf("cannot send alert %s", err)
		return err
	}

	log.Debug().Msgf("message to send is: %s", msg)
	input := &sns.PublishInput{}
	subject := fmt.Sprintf("[ALERT] rule %s is in alert state", n.Rule.Name)

	input.SetSubject(subject)
	input.SetMessage(msg)
	input.SetTopicArn(n.Rule.SnsTopic)

	result, err := client.Publish(input)
	if err != nil {
		log.Error().Msgf("sns publish error:", err)
		return err
	}
	log.Debug().Msgf("sns publish result for [%s]: %v+", n.Rule.Name, result)
	return nil
}

//SendOk - ok
func SendOk(wg *sync.WaitGroup, n *model.NotificationItem, t string, client *sns.SNS) error {
	defer wg.Done()
	log.Info().Msgf("sending OK for [%s]", n.Rule.Name)
	msg, err := RenderMessageTemplate(n, t)
	if err != nil {
		log.Error().Msgf("cannot render message template %s", err)
		return err
	}

	log.Debug().Msgf("message to send is: %s", msg)

	input := &sns.PublishInput{}
	subject := fmt.Sprintf("[OK] rule %s has recovered from alert state", n.Rule.Name)

	input.SetSubject(subject)
	input.SetMessage(msg)
	input.SetTopicArn(n.Rule.SnsTopic)

	result, err := client.Publish(input)
	if err != nil {
		log.Error().Msgf("sns publish error:", err)
		return err
	}
	log.Debug().Msgf("sns publish result for [%s]: %v+", n.Rule.Name, result)
	return nil
}

//RenderMessageTemplate - render message template and return text message
func RenderMessageTemplate(notification *model.NotificationItem, t string) (string, error) {
	tmpl := template.Must(template.New("sns").Parse(t))

	var b bytes.Buffer

	err := tmpl.Execute(&b, notification)
	if err != nil {
		log.Error().Err(err)
	}

	return b.String(), err
}
