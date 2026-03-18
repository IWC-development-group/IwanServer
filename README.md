# IwanServer

**Iwan server** is a generic server for [IwanClient](https://github.com/IWC-development-group/IwanClient) to store Markdown manuals.

## Iwan API specification
The Iwan API specification assumes the use of the HTTP protocol for transmitting requests. It focused on simplicity so you can make own IwanAPI server using this:

### Page request:
```
/?name=[namespace/<page name>]
```

### Response:
```json
{
	"status" : "response status (OK/ERR)",
	"name" : "actual page name (OR none)",
	"namespace" : "page namespace (OR none)",
	"content" : "page content (OR error description)"
}
```