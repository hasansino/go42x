package kwb

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "kwb"
	serverVersion = "0.2.0"
)

type MCPServer struct {
	service *Service
}

func NewMCPServer(service *Service) *MCPServer {
	return &MCPServer{
		service: service,
	}
}

func (s *MCPServer) Serve(_ context.Context) error {
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
	)

	searchTool := mcp.NewTool("search",
		mcp.WithDescription("Search the knowledge base"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("limit", mcp.Description("Maximum results (default: 10)")),
	)
	mcpServer.AddTool(searchTool, s.searchHandler)

	getFileTool := mcp.NewTool("get_file",
		mcp.WithDescription("Get full content of a specific file"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path")),
	)
	mcpServer.AddTool(getFileTool, s.getFileHandler)

	listFilesTool := mcp.NewTool("list_files",
		mcp.WithDescription("List all indexed files"),
		mcp.WithString("type", mcp.Description("Filter by type: code, documentation, config")),
	)
	mcpServer.AddTool(listFilesTool, s.listFilesHandler)

	return server.ServeStdio(mcpServer)
}

func (s *MCPServer) searchHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	limit := request.GetInt("limit", 10)

	results, err := s.service.Search(ctx, query, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search error: %v", err)), nil
	}

	output := fmt.Sprintf("Found %d results:\n\n", len(results))
	for i, result := range results {
		output += fmt.Sprintf("%d. %s (score: %.2f, type: %s)\n",
			i+1, result.Path, result.Score, result.Type)

		if result.Preview != "" {
			output += fmt.Sprintf("   Preview: %s\n", result.Preview)
		}
		output += "\n"
	}

	return mcp.NewToolResultText(output), nil
}

func (s *MCPServer) getFileHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	path := request.GetString("path", "")
	content, err := s.service.GetFile(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading file: %v", err)), nil
	}
	return mcp.NewToolResultText(content), nil
}

func (s *MCPServer) listFilesHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	fileType := request.GetString("type", "")

	files, err := s.service.ListFiles(ctx, fileType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error listing files: %v", err)), nil
	}

	output := fmt.Sprintf("Total files: %d\n\n", len(files))
	for _, file := range files {
		output += fmt.Sprintf("- %s\n", file)
	}

	return mcp.NewToolResultText(output), nil
}
