# JQ filter: press <CR> in normal mode to execute.

[.data.repository.releases.nodes[].releaseAssets.nodes[].name
| scan("[a-z0-9A-Z]+-[a-z0-9A-Z]+")] | sort | unique
