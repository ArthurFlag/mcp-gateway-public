package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/catalog"
	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/gateway/proxies"
)

func TestApplyConfigGrafana(t *testing.T) {
	catalogYAML := `
command:
  - --transport=stdio
secrets:
  - name: grafana.api_key
    env: GRAFANA_API_KEY
env:
  - name: GRAFANA_URL
    value: '{{grafana.url}}'
`
	configYAML := `
grafana:
  url: TEST
`
	secrets := map[string]string{
		"grafana.api_key": "API_KEY",
	}

	args, env := argsAndEnv(t, "grafana", catalogYAML, configYAML, secrets, nil)

	assert.Equal(t, []string{
		"run", "--rm", "-i", "--init", "--security-opt", "no-new-privileges", "--cpus", "1", "--memory", "2Gb", "--pull", "never",
		"-l", "docker-mcp=true", "-l", "docker-mcp-tool-type=mcp", "-l", "docker-mcp-name=grafana", "-l", "docker-mcp-transport=stdio",
		"-e", "GRAFANA_API_KEY", "-e", "GRAFANA_URL",
	}, args)
	assert.Equal(t, []string{"GRAFANA_API_KEY=API_KEY", "GRAFANA_URL=TEST"}, env)
}

func TestApplyConfigMongoDB(t *testing.T) {
	catalogYAML := `
secrets:
  - name: mongodb.connection_string
    env: MDB_MCP_CONNECTION_STRING
  `
	secrets := map[string]string{
		"mongodb.connection_string": "HOST:PORT",
	}

	args, env := argsAndEnv(t, "mongodb", catalogYAML, "", secrets, nil)

	assert.Equal(t, []string{
		"run", "--rm", "-i", "--init", "--security-opt", "no-new-privileges", "--cpus", "1", "--memory", "2Gb", "--pull", "never",
		"-l", "docker-mcp=true", "-l", "docker-mcp-tool-type=mcp", "-l", "docker-mcp-name=mongodb", "-l", "docker-mcp-transport=stdio",
		"-e", "MDB_MCP_CONNECTION_STRING",
	}, args)
	assert.Equal(t, []string{"MDB_MCP_CONNECTION_STRING=HOST:PORT"}, env)
}

func TestApplyConfigNotion(t *testing.T) {
	catalogYAML := `
secrets:
  - name: notion.internal_integration_token
    env: INTERNAL_INTEGRATION_TOKEN
    example: ntn_****
env:
  - name: OPENAPI_MCP_HEADERS
    value: '{"Authorization": "Bearer $INTERNAL_INTEGRATION_TOKEN", "Notion-Version": "2022-06-28"}'
  `
	secrets := map[string]string{
		"notion.internal_integration_token": "ntn_DUMMY",
	}

	args, env := argsAndEnv(t, "notion", catalogYAML, "", secrets, nil)

	assert.Equal(t, []string{
		"run", "--rm", "-i", "--init", "--security-opt", "no-new-privileges", "--cpus", "1", "--memory", "2Gb", "--pull", "never",
		"-l", "docker-mcp=true", "-l", "docker-mcp-tool-type=mcp", "-l", "docker-mcp-name=notion", "-l", "docker-mcp-transport=stdio",
		"-e", "INTERNAL_INTEGRATION_TOKEN", "-e", "OPENAPI_MCP_HEADERS",
	}, args)
	assert.Equal(t, []string{"INTERNAL_INTEGRATION_TOKEN=ntn_DUMMY", `OPENAPI_MCP_HEADERS={"Authorization": "Bearer ntn_DUMMY", "Notion-Version": "2022-06-28"}`}, env)
}

func TestApplyConfigMountAs(t *testing.T) {
	catalogYAML := `
volumes:
  - '{{hub.log_path|mount_as:/logs:ro}}'
  `
	configYAML := `
hub:
  log_path: /local/logs
`

	args, env := argsAndEnv(t, "hub", catalogYAML, configYAML, nil, nil)

	assert.Equal(t, []string{
		"run", "--rm", "-i", "--init", "--security-opt", "no-new-privileges", "--cpus", "1", "--memory", "2Gb", "--pull", "never",
		"-l", "docker-mcp=true", "-l", "docker-mcp-tool-type=mcp", "-l", "docker-mcp-name=hub", "-l", "docker-mcp-transport=stdio",
		"-v", "/local/logs:/logs:ro",
	}, args)
	assert.Empty(t, env)
}

func TestApplyConfigEmptyMountAs(t *testing.T) {
	catalogYAML := `
volumes:
  - '{{hub.log_path|mount_as:/logs:ro}}'
  `

	args, env := argsAndEnv(t, "hub", catalogYAML, "", nil, nil)

	assert.Equal(t, []string{
		"run", "--rm", "-i", "--init", "--security-opt", "no-new-privileges", "--cpus", "1", "--memory", "2Gb", "--pull", "never",
		"-l", "docker-mcp=true", "-l", "docker-mcp-tool-type=mcp", "-l", "docker-mcp-name=hub", "-l", "docker-mcp-transport=stdio",
	}, args)
	assert.Empty(t, env)
}

func TestApplyConfigMountAsReadOnly(t *testing.T) {
	catalogYAML := `
volumes:
  - '{{hub.log_path|mount_as:/logs:ro}}'
  `
	configYAML := `
hub:
  log_path: /local/logs
`

	args, env := argsAndEnv(t, "hub", catalogYAML, configYAML, nil, readOnly())

	assert.Equal(t, []string{
		"run", "--rm", "-i", "--init", "--security-opt", "no-new-privileges", "--cpus", "1", "--memory", "2Gb", "--pull", "never",
		"-l", "docker-mcp=true", "-l", "docker-mcp-tool-type=mcp", "-l", "docker-mcp-name=hub", "-l", "docker-mcp-transport=stdio",
		"-v", "/local/logs:/logs:ro",
	}, args)
	assert.Empty(t, env)
}

func argsAndEnv(t *testing.T, name, catalogYAML, configYAML string, secrets map[string]string, readOnly *bool) ([]string, []string) {
	t.Helper()

	clientPool := &clientPool{
		Options: Options{
			Cpus:   1,
			Memory: "2Gb",
		},
	}
	return clientPool.argsAndEnv(catalog.ServerConfig{
		Name:    name,
		Spec:    parseSpec(t, catalogYAML),
		Config:  parseConfig(t, configYAML),
		Secrets: secrets,
	}, readOnly, proxies.TargetConfig{})
}

func parseSpec(t *testing.T, contentYAML string) catalog.Server {
	t.Helper()
	var spec catalog.Server
	err := yaml.Unmarshal([]byte(contentYAML), &spec)
	require.NoError(t, err)
	return spec
}

func parseConfig(t *testing.T, contentYAML string) map[string]any {
	t.Helper()
	var config map[string]any
	err := yaml.Unmarshal([]byte(contentYAML), &config)
	require.NoError(t, err)
	return config
}

func readOnly() *bool {
	return boolPtr(true)
}

func boolPtr(b bool) *bool {
	return &b
}
