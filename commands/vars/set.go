package vars

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/daticahealth/cli/commands/services"
)

func CmdSet(svcName, defaultSvcID string, variables []string, iv IVars, is services.IServices) error {
	if svcName != "" {
		service, err := is.RetrieveByLabel(svcName)
		if err != nil {
			return err
		}
		if service == nil {
			return fmt.Errorf("Could not find a service with the label \"%s\". You can list services with the \"datica services\" command.", svcName)
		}
		defaultSvcID = service.ID
	}
	envVarsMap := make(map[string]string, len(variables))
	r := regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$")
	for _, envVar := range variables {
		pieces := strings.SplitN(envVar, "=", 2)
		if len(pieces) != 2 {
			return fmt.Errorf("Invalid variable format. Expected <key>=<value> but got %s", envVar)
		}
		name, value := pieces[0], pieces[1]
		if !r.MatchString(name) {
			return fmt.Errorf("Invalid environment variable name '%s'. Environment variable names must only contain letters, numbers, and underscores and must not start with a number.", name)
		}
		envVarsMap[name] = value
	}

	err := iv.Set(defaultSvcID, envVarsMap)
	if err != nil {
		return err
	}
	// TODO add in the service label in the redeploy example once we take in the service label in
	// this command
	logrus.Println("Set. For these environment variables to take effect, you will need to redeploy your service with \"datica redeploy\"")
	return nil
}

// Set adds a new environment variables or updates the value of an existing
// environment variables. Any changes to environment variables will not take
// effect until the service is redeployed by pushing new code or via
// `datica redeploy`.
func (v *SVars) Set(svcID string, envVarsMap map[string]string) error {
	b, err := json.Marshal(envVarsMap)
	if err != nil {
		return err
	}
	headers := v.Settings.HTTPManager.GetHeaders(v.Settings.SessionToken, v.Settings.Version, v.Settings.Pod, v.Settings.UsersID)
	resp, statusCode, err := v.Settings.HTTPManager.Post(b, fmt.Sprintf("%s%s/environments/%s/services/%s/env", v.Settings.PaasHost, v.Settings.PaasHostVersion, v.Settings.EnvironmentID, svcID), headers)
	if err != nil {
		return err
	}
	return v.Settings.HTTPManager.ConvertResp(resp, statusCode, nil)
}
