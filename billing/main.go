package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type timeWithZone struct {
	tzone *time.Location
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

	output, err := cw.GetMetricStatistics(&input)
	if err != nil {
		return "", err
	}

	return output.GoString(), err
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
