package main

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"gopkg.in/yaml.v2"
)

type serviceNames []string

type timeWithZone struct {
	tzone *time.Location
}

type output struct {
	metricStatistics cloudwatch.GetMetricStatisticsOutput
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

func handler(ctx context.Context) ([]string, error) {
	cfg := buildAWSConfig()
	s := session.New()
	cw := cloudwatch.New(s, cfg)

	inputs, err := buildInputs()
	if err != nil {
		return nil, err
	}
	outputs := make([]string, len(inputs))

	for _, input := range inputs {
		output, err := cw.GetMetricStatistics(input)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, output.GoString())
	}

	return outputs, err
}

func buildInputs() ([]*cloudwatch.GetMetricStatisticsInput, error) {
	f, err := os.Open("serviceName.yml")
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var names serviceNames
	err = yaml.Unmarshal(data, &names)

	inputs := make([]*cloudwatch.GetMetricStatisticsInput, len(names) + 1)
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

	inputs = append(inputs, &input)

	for _, serviceName := range names {
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

		inputs = append(inputs, &input)
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
	for _, v := range statistics {
		arr = append(arr, &v)
	}

	return arr
}
