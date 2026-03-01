package devtoolstestrunner

import (
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all test runner tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)
}
