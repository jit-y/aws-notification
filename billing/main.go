package main

import (
	"strings"
	"net/http"
	"fmt"
	"context"
	"io/ioutil"
	"os"
	"time"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"gopkg.in/yaml.v2"
	config "github.com/jit-y/aws-notification/config/billing"
)

type serviceNames []string

type timeWithZone struct {
	tzone *time.Location
}

type statisticsInput struct {
	name string
	input *cloudwatch.GetMetricStatisticsInput
}

type statisticsOutput struct {
	name string
	output *cloudwatch.GetMetricStatisticsOutput
}

func newTimeWithZone() *timeWithZone {
	t := timeWithZone{
		tzone: time.FixedZone("Asia/Tokyo", 9*60*60),
	}

	return &t
}

func (t *timeWithZone) beginningOfDay() time.Time {
	now := time.Now().UTC().In(t.tzone)

	year := now.Year()
	month := now.Month()
	day := now.Day()

	return time.Date(year, month, day-1, 0, 0, 0, 0, t.tzone)
}

func (t *timeWithZone) endOfDay() time.Time {
	now := time.Now().UTC().In(t.tzone)

	year := now.Year()
	month := now.Month()
	day := now.Day()

	return time.Date(year, month, day-1, 23, 59, 59, 59, t.tzone)
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) (string, error) {
	cfg := buildAWSConfig()
	s := session.New()
	cw := cloudwatch.New(s, cfg)

	inputs, err := buildInputs()
	if err != nil {
		return "", err
	}
	outputs := make([]statisticsOutput, len(inputs))

	for i, input := range inputs {
		output, err := cw.GetMetricStatistics(input.input)
		if err != nil {
			return "", err
		}

		outputs[i] = statisticsOutput{output: output, name: input.name}
	}

	attachment := buildAttatchment(outputs)
	reqBody, err := buildRequestBody(attachment)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	bodyReader := strings.NewReader(string(body))

	slackURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackURL == "" {
		return "", errors.New("SLACK_WEBHOOK_URL is not defined")
	}
	res, err := http.Post(slackURL, "application/json", bodyReader)
	if err != nil {
		return "", err
	}
	fmt.Println(res)

	return "ok", err
}

func buildInputs() ([]statisticsInput, error) {
	f, err := config.Assets.Open("/billing/servicename.yml")
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var names serviceNames
	err = yaml.Unmarshal(data, &names)

	inputs := make([]statisticsInput, len(names) + 1)
	t := newTimeWithZone()
	startTime := t.beginningOfDay()
	endTime := t.endOfDay()

	dimensions := []*cloudwatch.Dimension{
		buildDimension("Currency", "USD"),
	}
	statistics := buildStatistics("Average")

	input := cloudwatch.GetMetricStatisticsInput{}
	input.SetDimensions(dimensions)
	input.SetStartTime(startTime)
	input.SetEndTime(endTime)
	input.SetMetricName("EstimatedCharges")
	input.SetNamespace("AWS/Billing")
	input.SetPeriod(86400)
	input.SetStatistics(statistics)

	inputs[0] = statisticsInput{name: "Total", input: &input}

	for i, serviceName := range names {
		input := cloudwatch.GetMetricStatisticsInput{}
		dimensions := []*cloudwatch.Dimension{
			buildDimension("Currency", "USD"),
			buildDimension("ServiceName", serviceName),
		}
		input.SetDimensions(dimensions)
		input.SetStartTime(startTime)
		input.SetEndTime(endTime)
		input.SetMetricName("EstimatedCharges")
		input.SetNamespace("AWS/Billing")
		input.SetPeriod(86400)
		input.SetStatistics(statistics)

		inputs[i+1] = statisticsInput{name: serviceName, input: &input}
	}

	return inputs, nil
}

func buildAWSConfig() *aws.Config {
	endpoint := os.Getenv("CW_ENDPOINT")
	region := os.Getenv("CW_REGION")

	return &aws.Config{
		Endpoint: &endpoint,
		Region:   &region,
	}
}

func buildDimension(name, value string) *cloudwatch.Dimension {
	dimension := cloudwatch.Dimension{}
	dimension.SetName(name)
	dimension.SetValue(value)

	return &dimension
}

func buildStatistics(statistics ...string) []*string {
	arr := make([]*string, len(statistics))
	for i, v := range statistics {
		arr[i] = &v
	}

	return arr
}

type reqAttachment struct {
	Fallback string `json:"fallback"`
	Pretext string	`json:"pretext"`
	Color string `json:"color"`
	Fields []*reqField `json:"fields"`
}

type reqField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool `json:"short"`
}

func buildAttatchment(statisticsOutputs []statisticsOutput) *reqAttachment {
	attachment := reqAttachment{Fallback: "oops", Pretext: "", Color: "good"}
	fields := make([]*reqField, len(statisticsOutputs))
	for i, so := range statisticsOutputs {
		var point float64
		for _, d := range so.output.Datapoints {
			avg := d.Average
			if avg != nil {
				point += *avg
			}
		}
		fields[i] = &reqField{
			Title: so.name,
			Value: fmt.Sprintf("%f USD", point),
			Short: true,
		}
	}
	attachment.Fields = fields

	return &attachment
}

type reqBody struct {
	Channel string `json:"channel"`
	Attachments []*reqAttachment `json:"attachments"`
}

func buildRequestBody(attachments ...*reqAttachment) (*reqBody, error) {
	channelName := os.Getenv("SLACK_CHANNEL_NAME")
	if channelName == "" {
		return nil, errors.New("SLACK_CHANNEL_NAME is not defined")
	}

	body := reqBody{
		Channel: channelName,
		Attachments: attachments,
	}

	return &body, nil
}
