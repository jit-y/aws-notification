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

type service struct {
	title   string
	billing int
}

func handler(ctx context.Context) (string, error) {
	endpoint := os.Getenv("CW_END_POINT")
	region := os.Getenv("CW_REGION")

	cfg := aws.Config{
		Endpoint: &endpoint,
		Region:   &region,
	}
	s := session.New()
	cw := cloudwatch.New(s, &cfg)

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().UTC().In(jst)

	year := now.Year()
	month := now.Month()
	day := now.Day()
	startTime := time.Date(year, month, day-1, 0, 0, 0, 0, jst)
	endTime := time.Date(year, month, day-1, 23, 59, 59, 0, jst)

	currency := cloudwatch.Dimension{}
	currency.SetName("Currency")
	currency.SetValue("USD")
	currency.SetName("ServiceName")
	currency.SetValue("AmazonEC2")
	dimensions := []*cloudwatch.Dimension{
		&currency,
	}
	avg := "Average"
	statistics := []*string{
		&avg,
	}

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

func main() {
	lambda.Start(handler)
}
