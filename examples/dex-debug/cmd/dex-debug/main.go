package main

import (
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/kotaicode/pulumi-provider-dex/examples/dex-debug/internal/commands"
)

var cli struct {
	Host string `help:"Dex gRPC host:port" default:"localhost:5557" short:"H"`

	Verify             commands.VerifyCmd             `cmd:"" help:"List all clients and connectors in Dex" aliases:"list"`
	Cleanup            commands.CleanupCmd            `cmd:"" help:"Clean up test clients and connectors (excluding static ones)"`
	TestDelete         commands.TestDeleteCmd         `cmd:"" help:"Test deleting a specific client by ID"`
	TestDeleteDirect   commands.TestDeleteDirectCmd   `cmd:"" help:"Test DeleteClient API with a test client (creates, deletes, verifies)"`
	TestDeleteMyWebApp commands.TestDeleteMyWebAppCmd `cmd:"" help:"Test DeleteClient API with 'my-web-app' client"`
	TestVerification   commands.TestVerificationCmd   `cmd:"" help:"Test delete verification logic with 'my-web-app' client"`
}

func injectHost(cmd interface{}, host string) {
	v := reflect.ValueOf(cmd).Elem()
	baseCmd := v.FieldByName("BaseCmd")
	if baseCmd.IsValid() {
		hostField := baseCmd.FieldByName("Host")
		if hostField.IsValid() && hostField.CanSet() {
			hostField.SetString(host)
		}
	}
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("dex-debug"),
		kong.Description("Debugging and testing tool for Dex gRPC API"),
	)

	// Inject host into the selected command
	if cmd := ctx.Selected(); cmd != nil {
		injectHost(cmd.Target.Addr().Interface(), cli.Host)
	}

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
