package cmd

// Copyright ©2016 Fabian Stegemann
//
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"fmt"
	"github.com/adjust/rmq"
	"strings"
)

type DeploymentEventConsumer struct {
	name string
}

func (consumer *DeploymentEventConsumer) Consume(delivery rmq.Delivery) {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(delivery.Payload()), &event); err != nil {
		delivery.Reject()

		log.WithFields(log.Fields{
			"error":   err,
			"payload": delivery.Payload(),
		}).Error("Could not parse payload.")

		return
	}

	var deploymentID string
	var secretsVolumeName string
	var cacheVolumeName string

	var deploymentIDBuffer bytes.Buffer
	var secretsVolumeNameBuffer bytes.Buffer
	var cacheVolumeNameBuffer bytes.Buffer

	if err := deploymentIDTemplate.Execute(&deploymentIDBuffer, event); err != nil {
		delivery.Reject()

		log.WithFields(log.Fields{
			"consumer": consumer.name,
			"error":    err,
		}).Error("Could not render deploymentID template.")

		return
	}

	if err := cacheVolumeNameTemplate.Execute(&cacheVolumeNameBuffer, event); err != nil {
		delivery.Reject()

		log.WithFields(log.Fields{
			"consumer": consumer.name,
			"error":    err,
		}).Error("Could not render cache volume name template.")

		return
	}

	if err := secretsVolumeNameTemplate.Execute(&secretsVolumeNameBuffer, event); err != nil {
		delivery.Reject()

		log.WithFields(log.Fields{
			"consumer": consumer.name,
			"error":    err,
		}).Error("Could not render secrets volume name template.")

		return
	}

	deploymentID = deploymentIDBuffer.String()
	secretsVolumeName = secretsVolumeNameBuffer.String()
	cacheVolumeName = cacheVolumeNameBuffer.String()

	script := fmt.Sprintf(
		`#!/bin/sh
stdin=$(</dev/stdin)
docker volume create %s
echo "$stdin" | docker run --rm -i -v %s:/var/cache/deployment alpine:3.4 dd of=/var/cache/deployment/deployment-event.json
SECRETS_VOLUME_NAME=%s DEPLOYMENT_CACHE_VOLUME_NAME=%s docker-compose --file docker-compose.deploy.yml --project-name %s up -d
docker wait %s_deploy
`,
		cacheVolumeName,
		cacheVolumeName,
		secretsVolumeName,
		cacheVolumeName,
		cacheVolumeName,
	)
	scriptRunnerCommand := exec.Command(
		"/bin/sh",
		script,
	)
	scriptRunnerCommand.Stdout = os.Stdout
	scriptRunnerCommand.Stderr = os.Stderr
	scriptRunnerCommand.Stdin = strings.NewReader(delivery.Payload())
	if err := scriptRunnerCommand.Run(); err != nil {
		delivery.Reject()

		log.WithFields(log.Fields{
			"consumer":       consumer.name,
			"error":          err,
			"worker-image":   workerImageName,
			"deployment":     deploymentID,
			"raw-payload":    delivery.Payload(),
			"parsed-payload": event,
			"cache-volume":   cacheVolumeName,
		}).Error("Failed to run deployment worker.")

		return
	}

	// ------------------------------

	// if all went well
	delivery.Ack()
}
