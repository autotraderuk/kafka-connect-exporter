package aws

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jackpal/gateway"
	"github.com/pkg/errors"
	"github.com/zenreach/go-aws/ecsagent"
	"github.com/zenreach/hatchet"
)

const (
	// RegionField is the name of the field containing the AWS region.
	RegionField = "ec2_region"
	// InstanceField is the name of the field containing the EC2 instance ID.
	InstanceField = "ec2_instance"
	// TaskARNField is the name of the field containing the ECS task ARN.
	TaskARNField = "ecs_task_arn"
	// TaskFamilyField is the name of the field containing the ECS task family.
	TaskFamilyField = "ecs_task_family"
	// TaskVersionField is the name of the field containing the ECS task version.
	TaskVersionField = "ecs_task_version"
)

const (
	ecsAgentPort = 51678
)

// Session is used by EC2Info to retrieve details from AWS.
var Session = session.Must(session.NewSession())

// EC2Info adds application information to a log in a similar manner to
// logutil.AppInfo. It retrieves the hostname from EC2 rather than
// os.Hostname(). It will also attempt to retrieve the ECS service and task
// names if it has permission.
//
// The following fields are added to the log if available:
// - ec2_region: The region in which the instance is running.
// - ec2_instance: The ID of the instance.
// - ecs_task: The ECS task ARN.
// - ecs_taskdef: The ECS task definition.
//
// The ECS info requires that the ecs-agent's HTTP server be accessible.
// Regular app info is appended if the EC2 metadata endpoint is unavailable.
func EC2Info(logger hatchet.Logger) hatchet.Logger {
	fields := hatchet.L{
		hatchet.PID:      os.Getpid(),
		hatchet.Process:  getProcess(),
		hatchet.Hostname: getHostname(),
	}

	setEC2Fields(fields)
	setECSFields(fields)
	return hatchet.Fields(logger, fields, true)
}

func setEC2Fields(fields hatchet.L) {
	client := ec2metadata.New(Session)
	if !client.Available() {
		fmt.Fprintf(os.Stderr, "ec2 metadata endpoint unavailable\n")
		return
	}

	if identity, err := client.GetInstanceIdentityDocument(); err != nil {
		fmt.Fprintf(os.Stderr, "ec2 instance identity unavailable: %s\n", err)
	} else {
		fields[RegionField] = identity.Region
		fields[InstanceField] = identity.InstanceID
	}

	if hostname, err := client.GetMetadata("local-hostname"); err != nil {
		fmt.Fprintf(os.Stderr, "ec2 hostname unavailable: %s\n", err)
	} else if hostname == "" {
		fmt.Fprint(os.Stderr, "ec2 hostname unavailable\n")
	} else {
		fields[hatchet.Hostname] = hostname
	}
}

func setECSFields(fields hatchet.L) {
	gw, err := gateway.DiscoverGateway()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ecs info unavailable: %s\n", err)
		return
	}
	id, err := getContainerID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ecs info unavailable: %s\n", err)
		return
	}
	task, err := ecsagent.NewClient(gw.String()).TaskByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ecs info unavailable: unable to query ecs agent: %s\n", err)
		return
	}
	if task.ARN == "" {
		fmt.Fprintf(os.Stderr, "ecs info unavailable: container %s not found\n", id)
		return
	}
	fields[TaskARNField] = task.ARN
	fields[TaskFamilyField] = task.Family
	fields[TaskVersionField] = task.Version
}

func getProcess() string {
	if len(os.Args) == 0 || os.Args[0] == "" {
		return ""
	}
	return path.Base(os.Args[0])
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func getContainerID() (string, error) {
	file, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", errors.Wrap(err, "unable to open cgroup file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "docker") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}
		paths := strings.Split(parts[2], "/")
		if len(paths) == 0 {
			continue
		}
		id := strings.TrimSpace(paths[len(paths)-1])
		id = strings.TrimPrefix(id, "docker-")
		return strings.TrimSuffix(id, ".scope"), nil
	}
	return "", errors.New("docker cgroups not found")
}
