# Pull Request: Add Simplified Azure Advisor Prompt for AKS-MCP

## Overview
This PR adds a simplified Azure Advisor prompt file for the AKS-MCP server, focusing on high-level implementation instructions for Azure Advisor recommendation capabilities.

## Changes Made
- ✅ **Added** `prompts/azure-advisor.md` - Simplified Azure Advisor implementation prompt
- ✅ **Updated** `prompts/README.md` - Added documentation for the new prompt file

## Features Added

### Azure Advisor Tool: `az_advisor_recommendation`
A unified tool that provides Azure Advisor recommendations specifically for AKS clusters and related resources.

**Operations Supported:**
- `list` - Return AKS-related recommendations with basic details
- `details` - Get comprehensive information for a specific recommendation  
- `report` - Generate summary report of AKS recommendations by category

**Key Parameters:**
- `operation` (required): `list`, `details`, or `report`
- `subscription_id` (required): Azure subscription ID
- `resource_group` (optional): Filter by resource group
- `cluster_names` (optional): Array of AKS cluster names to filter
- `category` (optional): `Cost`, `HighAvailability`, `Performance`, `Security`
- `severity` (optional): `High`, `Medium`, `Low`
- `recommendation_id` (optional): Required for `details` operation

## Implementation Approach
- **Existing Executor Integration**: Uses existing executor from `internal/az/executor.go` for Azure CLI commands
- **AKS Focus**: Filters recommendations to only include AKS-related resources
- **Simplified Instructions**: High-level implementation guidance without excessive technical details
- **Error Handling**: Graceful error handling with meaningful messages
- **Structured Output**: Standard JSON output format

## File Structure
```
prompts/
├── README.md                    # Updated documentation
├── monitoringservice.md        # Existing monitoring service integration
├── azure-diagnostics.md        # Existing diagnostics tools specifications
└── azure-advisor.md            # NEW: Azure Advisor recommendations integration
```

## Access Levels
- **Readonly**: All operations supported
- **Readwrite**: Enhanced filtering options
- **Admin**: Same as readwrite (no admin-specific operations)

## Testing Considerations
- Tool should integrate with existing AKS-MCP server architecture
- Should follow established error handling patterns
- Should use standard JSON output format
- Should support all configured access levels

## Documentation Updates
- Updated `prompts/README.md` to include the new Azure Advisor prompt file
- Maintained consistency with existing prompt file documentation

## Breaking Changes
None - this is a new addition that doesn't affect existing functionality.

## Related Issues
- Implements Azure Advisor recommendation capabilities for AKS clusters
- Provides simplified, high-level implementation instructions
- Focuses on AKS-specific use cases and scenarios

## Checklist
- [x] Code follows project conventions
- [x] Documentation has been updated
- [x] Changes are focused and minimal
- [x] Commit message follows conventional format
- [x] No breaking changes introduced
- [x] Files are properly structured and organized

## Next Steps
After this PR is merged, implementers can use the `azure-advisor.md` prompt to:
1. Add tool registration in `internal/server/server.go`
2. Create handlers in `internal/azure/advisor/` directory
3. Use existing executor from `internal/az/executor.go` for Azure CLI commands
4. Follow existing error handling patterns
5. Use standard JSON output format

## Review Notes
This PR focuses on providing clear, concise implementation guidance for Azure Advisor integration without overwhelming technical details. The prompt is designed to be actionable and easy to follow for developers implementing the feature.
