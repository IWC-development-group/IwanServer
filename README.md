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
---
> [!CAUTION]
> If no namespace specified for the page **it's namespace needs to be named as "global"!**

# Usage
Host the server with specified port:
```
iwans serve -p [port]
```

Add manuals to the server's database:
```
iwans index [manuals_directory] -n [namespace_name]
```
If no namespace specified it will use "global" by default.

Type `iwans -h` for details.