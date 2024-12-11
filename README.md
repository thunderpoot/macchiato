# Macchiato

![logo](macchiato-logo.webp)

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

**Macchiato** is a feature-packed Markdown-to-HTML converter written in Go designed to make your documents shine. Create blog posts, documentation, or just explore the magic of Markdown. Macchiato transforms your text into elegant, web-ready HTML.

Nearly everyone has written a Markdown converter at some point. This project started as a tool for a very specific purpose, but it turned out surprisingly well, so I decided to share it here.

## 🚀 Features
- **Intuitive Markdown Parsing**: Supports the stuff you'd imagine it supports, like headings, bold text, italics, links, images, code blocks, but also some other stuff.
- **Customisable CSS Styling**: Specify your CSS styles with the `--style` option or let Macchiato use its default style.
- **Footnotes and Highlights**: Easily add footnotes and highlighted text to your documents.
- **Dynamic Lists and Tables**: Supports ordered and unordered lists, as well as Markdown tables, obviously.
- **Code Syntax Highlighting**: Formatted code blocks with a copy button, (powered by [Prism](https://prismjs.com/)†.)
- **Extensible HTML Template**: Use the `main.html` template to customise the look of your output.

<sub>_† Feel free to use some other highlighter, just edit the `main.html` template file._</sub>

## Basic Usage

```bash
go run macchiato.go <input_file> [output_file]
```

- `<input_file>`: Path to your Markdown file.
- `[output_file]`: (Optional) Path to save the resulting HTML. If not specified, the HTML is printed to stdout.

### With Custom Styles

```
go run macchiato.go <input_file> [output_file] --style=<css_file>
```


## Markdown Features Supported

-   Headings: `#`, `##`, `###`, etc.
-   Text Formatting: Bold (`**text**`), Italics (`*text*`), Strikethrough (`~~text~~`).
-   Links and Images: `[Link](url)` and `![Alt Text](image_url)`.
-   Lists: Ordered (`1. Item`) and Unordered (`- Item`).
-   Code Blocks: Indented or fenced with \`\`\`.
-   Tables: Markdown table syntax with headers and rows.
-   Footnotes: `[Text][^1]` with `[^1]`: Footnote text.
-   Custom Containers: `:::` note for callout boxes.


## License

This project is licensed under the MIT License. Use it freely, modify it, and share it with others.


## Contributors

Contributions and feedback are always welcome.

## References

- [Markdown on Wikipedia](https://en.wikipedia.org/wiki/Markdown)
- [RFC 7764](https://datatracker.ietf.org/doc/html/rfc7764.html)
- [RFC 7763](https://datatracker.ietf.org/doc/html/rfc7763.html)
