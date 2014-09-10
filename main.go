package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/goamz/autoscaling"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	// . "github.com/motain/gocheck"
)

type Ec2 struct {
	conn *ec2.EC2
	asg  *autoscaling.AutoScaling
}

func (s *Ec2) imageFindByName(nameish string) (*ec2.ImagesResp, error) {
	filter := ec2.NewFilter()
	filter.Add("architecture", "x86_64")
	filter.Add("root-device-type", "ebs")
	filter.Add("name", "app-mw*")
	imagesResp, err := s.conn.ImagesByOwners([]string{}, []string{"self"}, filter)
	if err != nil {
		log.Print("Error: ", err)
		return nil, err
	}

	return imagesResp, nil
}

func main() {

	auth, err := aws.SharedAuth()
	if err != nil {
		log.Fatal("Error: ", err)
	}

	s := &Ec2{
		conn: ec2.New(auth, aws.USEast),
		asg:  autoscaling.New(auth, aws.USEast)}
	if s.conn == nil || s.asg == nil {
		log.Fatal("Error connecting to ec2.")
	}

	imagesResp, err := s.imageFindByName("app-mw*")
	if err != nil {
		log.Fatal("Error: ", err)
	}

	if b, err := json.MarshalIndent(imagesResp, "", " "); err == nil {
		fmt.Print(string(b))
	}

	descLaunchConfigs, err := s.asg.DescribeLaunchConfigurations(
		&autoscaling.DescribeLaunchConfigurations{
			Names: []string{"qa-analytics-lc-v002", "awseb-e-dicxfpesmm-stack-AWSEBAutoScalingLaunchConfiguration-1J8QOTEP7VCAL"},
		})

	if err != nil {
		log.Fatal("Error: ", err)
	}
	// var v autoscaling.LaunchConfiguration
	for _, v := range descLaunchConfigs.LaunchConfigurations {
		//var imageId string
		v.UserData = make([]byte, 0)
		//imageId = v.ImageId
		if b, err := json.MarshalIndent(v, "", " "); err == nil {
			fmt.Print(string(b))
		}
	}

	if os.Getenv("AWS_TEST") != "" {

		options := autoscaling.CreateAutoScalingGroup{
			AvailZone:               []string{"us-east-1a"},
			DefaultCooldown:         30,
			HealthCheckGracePeriod:  30,
			LaunchConfigurationName: "amigo-kaios-launch-v0001",
			SetMinSize:              true,
			SetMaxSize:              true,
			SetDesiredCapacity:      true,
			MinSize:                 0,
			MaxSize:                 0,
			DesiredCapacity:         0,
			HealthCheckType:         "EC2",
			TerminationPolicies:     []string{"ClosestToNextInstanceHour", "OldestInstance"},
			Name:                    "test1",
			Tags: []autoscaling.Tag{
				autoscaling.Tag{
					Key:   "foo",
					Value: "bar",
				},
			},
		}

		resp, err := s.asg.CreateAutoScalingGroup(&options)
		if err == nil {
			fmt.Printf("%#v\n", resp)
		} else {
			fmt.Printf("%#v\n", err)
		}

		resp, err = s.asg.DeleteAutoScalingGroup(&autoscaling.DeleteAutoScalingGroup{Name: "test1"})
		if err == nil {
			log.Print("Deleted asg", resp)
		} else {
			log.Print("Error deleting asg", err)
		}
	}
}
