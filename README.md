# IwanServer

**Iwan server** is a generic server for [IwanClient](https://github.com/IWC-development-group/IwanClient) to store Markdown manuals.

## Iwan API specification
The Iwan API specification assumes the use of the HTTP protocol for transmitting requests. It focused on simplicity so you can make own IwanAPI server using this:

### Page querying:
#### Request:
```
/?name=[namespace/<page name>]
```

#### Response:
```json
{
	"status" : "response status (OK/ERR)",
	"name" : "actual page name (OR none)",
	"namespace" : "page namespace (OR none)",
	"content" : "page content (OR error description)"
}
```

---
> [!IMPORTANT]
> If no namespace specified for the page **it's namespace needs to be named as "global"!**
---

### Listing pages in the specified namespace
#### Request:
```
/pages?namespace=[namespace name]
```

#### Response:
```json
{
	"status": "response status (OK/ERR)",
	"namespace": "actual namespace",
	"pages": ["page1 (OR error description)", "page2", "page3", ...]
}
```

### Listing all namespaces that contains at least one page
#### Request:
```
/namespaces
```

#### Response:
```json
{
	"status": "response status (OK/ERR)",
	"namespaces": ["namespace1 (OR error description)", "namespace2", "namespace3", ...]
}
```

# Usage
## Server
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

## Hints
To automatically determine namespaces of indexed pages, you can use hints.

**Usage:**
1) Place a file `hint.iwan` in the directory that you want to associate with a specific namespace.
2) Open the hint file and write namespace name inside, then save.
3) Run `iwan index [manuals_directory]`.

**Example:**
```
my-docs/opengl
├───es3.1
│   ├───hint.iwan <-- contains es3.1
│	└───...pages
├───gl2
│   ├───hint.iwan <-- contains gl2
│	└───...pages
└───gl4
    ├───hint.iwan <-- contains gl4
 	└───...pages
```

---
> [!CAUTION]
> Note that hinted indexing requires the additional directory pass to collect all hints! If you don't want it, use `-f` flag to force a specific namespace. 
---

## Converter
This project also includes a converter utility to convert HTML pages to Markdown:
```
iwanc [source] [destination]
```
It will recursively scan the `source` directory tree for HTML or XHTML files, convert them to Markdown, and place them to the `destination` path, preserving the source directory structure.