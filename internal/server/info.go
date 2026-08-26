package server

import (
	"context"

	"github.com/granitebps/threads-mcp/internal/provider"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerInfoOutput struct {
	ServerName        string   `json:"server_name"`
	ServerVersion     string   `json:"server_version"`
	Commit            string   `json:"commit"`
	BuildDate         string   `json:"build_date"`
	SDKVersion        string   `json:"sdk_version"`
	ProtocolVersion   string   `json:"protocol_version"`
	ProviderName      string   `json:"provider_name"`
	ProviderVersion   string   `json:"provider_version"`
	ProviderMode      string   `json:"provider_mode"`
	ProviderUserAgent string   `json:"provider_user_agent"`
	SupportedTools    []string `json:"supported_tools"`
	ExcludedTools     []string `json:"excluded_tools"`
	SupportedInputs   []string `json:"supported_inputs"`
	KnownLimitations  []string `json:"known_limitations"`
}

func addServerInfoTool(server *mcp.Server, p provider.Provider, build BuildInfo) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server_info",
		Description: "Report threads-mcp versions, access mode, supported tools, and known public-data limits.",
		Annotations: readAnnotations(),
	}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, ServerInfoOutput, error) {
		info := p.Info()
		return nil, ServerInfoOutput{
			ServerName:        "threads-mcp",
			ServerVersion:     build.Version,
			Commit:            build.Commit,
			BuildDate:         build.Date,
			SDKVersion:        build.SDKVersion,
			ProtocolVersion:   build.Protocol,
			ProviderName:      info.Name,
			ProviderVersion:   info.Version,
			ProviderMode:      info.Mode,
			ProviderUserAgent: info.UserAgent,
			SupportedTools:    append([]string(nil), expectedToolNames...),
			ExcludedTools:     []string{"get_profile_replies"},
			SupportedInputs:   append([]string(nil), info.SupportedInputs...),
			KnownLimitations:  append([]string(nil), info.KnownLimitations...),
		}, nil
	})
}
