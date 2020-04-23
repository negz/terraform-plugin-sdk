package plugin

import (
	"context"
	"errors"
	"log"
	"net/rpc"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
)

var (
	_ plugin.GRPCPlugin = (*gRPCProviderPlugin)(nil)
	_ plugin.Plugin     = (*gRPCProviderPlugin)(nil)
)

// gRPCProviderPlugin implements plugin.GRPCPlugin and plugin.Plugin for the go-plugin package.
// the only real implementation is GRPCSServer, the other methods are only satisfied
// for compatibility with go-plugin
type gRPCProviderPlugin struct {
	GRPCProvider func() proto.ProviderServer
}

func (p *gRPCProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, errors.New("terraform-plugin-sdk only implements grpc servers")
}

func (p *gRPCProviderPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, errors.New("terraform-plugin-sdk only implements grpc servers")
}

func (p *gRPCProviderPlugin) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
	return nil, errors.New("terraform-plugin-sdk only implements grpc servers")
}

func (p *gRPCProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterProviderServer(s, p.GRPCProvider())
	return nil
}

// closing the grpc connection is final, and terraform will call it at the end of every phase.
func (p *GRPCProvider) Close() error {
	log.Printf("[TRACE] GRPCProvider: Close")

	// Make sure to stop the server if we're not running within go-plugin.
	if p.TestServer != nil {
		p.TestServer.Stop()
	}

	// Check this since it's not automatically inserted during plugin creation.
	// It's currently only inserted by the command package, because that is
	// where the factory is built and is the only point with access to the
	// plugin.Client.
	if p.PluginClient == nil {
		log.Println("[DEBUG] provider has no plugin.Client")
		return nil
	}

	if os.Getenv("TF_PROVIDER_SOFT_STOP") != "" {
		log.Println("[DEBUG] detected we want a soft stop of providers")
		// TODO: ideally, we'd gate this on the provider, so only the
		// provider the user specifies exhibits this behavior, and the
		// rest get the usual behavior. Unfortunately, we don't
		// actually have access at this point to the terraform.Addr of
		// the provider we're talking to, so we have to just have this
		// behavior for everyone.
		c, err := p.PluginClient.Client()
		if err != nil {
			log.Println("[ERROR] can't obtain client for provider, killing process instead of server")
		} else {
			log.Println("[DEBUG] calling Stop instead of Kill on provider")
			return c.Close()
		}
	}

	p.PluginClient.Kill()
	return nil
}
