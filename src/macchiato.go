package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "html"
)

type Rule struct {
    pattern *regexp.Regexp
    action  func([]string) string
}

type ListType int

const (
    Unordered ListType = iota
    Ordered
)

type ListItem struct {
    Type    ListType
    Number  int
    Indent  int
    Content string
}

type Macchiato struct {
    rules        []Rule
    footnotes    map[string]string
    currentLists []ListItem
}

func NewMacchiato() *Macchiato {
    m := &Macchiato{
        footnotes: make(map[string]string),
    }
    m.initRules()
    return m
}

func (m *Macchiato) initRules() {
    m.rules = []Rule{
        {regexp.MustCompile(`^# (.+)$`), func(groups []string) string {
            id := m.generateID(groups[1])
            return fmt.Sprintf("<h1 id=\"%s\"><a href=\"#%s\">%s</a></h1>", id, id, groups[1])
        }},
        {regexp.MustCompile(`^## (.+)$`), func(groups []string) string {
            id := m.generateID(groups[1])
            return fmt.Sprintf("<h2 id=\"%s\"><a href=\"#%s\">%s</a></h2>", id, id, groups[1])
        }},
        {regexp.MustCompile(`^### (.+)$`), func(groups []string) string {
            id := m.generateID(groups[1])
            return fmt.Sprintf("<h3 id=\"%s\"><a href=\"#%s\">%s</a></h3>", id, id, groups[1])
        }},
        {regexp.MustCompile(`^#### (.+)$`), func(groups []string) string {
            id := m.generateID(groups[1])
            return fmt.Sprintf("<h4 id=\"%s\"><a href=\"#%s\">%s</a></h4>", id, id, groups[1])
        }},
        {regexp.MustCompile(`^##### (.+)$`), func(groups []string) string {
            id := m.generateID(groups[1])
            return fmt.Sprintf("<h5 id=\"%s\"><a href=\"#%s\">%s</a></h5>", id, id, groups[1])
        }},
        {regexp.MustCompile(`\*\*(.+?)\*\*`), func(groups []string) string { return fmt.Sprintf("<strong>%s</strong>", groups[1]) }},
        {regexp.MustCompile(`\*(.+?)\*`), func(groups []string) string { return fmt.Sprintf("<em>%s</em>", groups[1]) }},
        {regexp.MustCompile("`(.+?)`"), func(groups []string) string { return fmt.Sprintf("<code>%s</code>", groups[1]) }},
        {regexp.MustCompile(`!\[(.+?)\]\((.+?)\)`), func(groups []string) string { return fmt.Sprintf("<img src=\"%s\" alt=\"%s\">", groups[2], groups[1]) }},
        {regexp.MustCompile(`\[(.+?)\]\((.+?)\)`), func(groups []string) string { return fmt.Sprintf("<a href=\"%s\">%s</a>", groups[2], groups[1]) }},
        {regexp.MustCompile(`==(.+?)==`), func(groups []string) string { return fmt.Sprintf("<mark>%s</mark>", groups[1]) }},
        {regexp.MustCompile(`~~(.+?)~~`), func(groups []string) string { return fmt.Sprintf("<del>%s</del>", groups[1]) }},
        {regexp.MustCompile(`\[\^(.+?)\]`), m.handleFootnoteReference},
        {regexp.MustCompile(`^\[\^(.+?)\]:\s*(.+)$`), m.handleFootnoteDefinition},
    }
}

func (m *Macchiato) generateID(text string) string {
    id := strings.ToLower(text)
    id = strings.ReplaceAll(id, " ", "-")
    id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")
    id = strings.Trim(id, "-")
    return id
}

func (m *Macchiato) Parse(markdown string) string {
    lines := strings.Split(markdown, "\n")
    var html strings.Builder
    var codeBlock strings.Builder
    var customContainer strings.Builder
    inCodeBlock := false
    inCustomContainer := false
    codeLanguage := ""
    containerType := ""
    inTable := false
    var tableRows []string

    for _, line := range lines {
        trimmedLine := strings.TrimLeft(line, " ")
        indent := len(line) - len(trimmedLine)

        if strings.HasPrefix(line, "```") {
            if inCodeBlock {
                html.WriteString(m.formatCodeBlock(codeBlock.String(), codeLanguage))
                codeBlock.Reset()
                inCodeBlock = false
                codeLanguage = ""
            } else {
                inCodeBlock = true
                codeLanguage = strings.TrimSpace(strings.TrimPrefix(line, "```"))
            }
            continue
        }

        if inCodeBlock {
            codeBlock.WriteString(line + "\n")
            continue
        }

        if strings.HasPrefix(line, ":::") {
            if inCustomContainer {
                html.WriteString(m.formatCustomContainer(customContainer.String(), containerType))
                customContainer.Reset()
                inCustomContainer = false
                containerType = ""
            } else {
                inCustomContainer = true
                containerType = strings.TrimSpace(strings.TrimPrefix(line, ":::"))
            }
            continue
        }

        if inCustomContainer {
            customContainer.WriteString(line + "\n")
            continue
        }

        if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") || regexp.MustCompile(`^\d+\. `).MatchString(trimmedLine) {
            m.handleListItem(trimmedLine, indent)
            continue
        }

        if len(m.currentLists) > 0 {
            html.WriteString(m.renderLists())
        }

        if strings.HasPrefix(trimmedLine, "|") && strings.HasSuffix(trimmedLine, "|") {
            if !inTable {
                inTable = true
                tableRows = []string{}
            }
            tableRows = append(tableRows, trimmedLine)
        } else if inTable {
            inTable = false
            html.WriteString(m.handleTable(tableRows))
            tableRows = []string{}
        } else {
            parsedLine := m.parseLine(line)
            if parsedLine != "" {
                html.WriteString("<p>" + parsedLine + "</p>\n")
            }
        }
    }

    if inTable {
        html.WriteString(m.handleTable(tableRows))
    }

    if len(m.currentLists) > 0 {
        html.WriteString(m.renderLists())
    }

    html.WriteString(m.renderFootnotes())

    return html.String()
}

func (m *Macchiato) parseLine(line string) string {
    for _, rule := range m.rules {
        if rule.pattern.MatchString(line) {
            line = rule.pattern.ReplaceAllStringFunc(line, func(match string) string {
                groups := rule.pattern.FindStringSubmatch(match)
                return rule.action(groups)
            })
        }
    }
    return line
}

func (m *Macchiato) handleListItem(line string, indent int) {
    listType := Unordered
    if regexp.MustCompile(`^\d+\. `).MatchString(line) {
        listType = Ordered
    }

    content := regexp.MustCompile(`^(\d+\.|-|\*)\s*`).ReplaceAllString(line, "")
    item := ListItem{Type: listType, Indent: indent, Content: content}

    if listType == Ordered {
        item.Number = m.getNextNumber(indent)
    }

    m.currentLists = append(m.currentLists, item)
}

func (m *Macchiato) getNextNumber(indent int) int {
    for i := len(m.currentLists) - 1; i >= 0; i-- {
        if m.currentLists[i].Indent == indent && m.currentLists[i].Type == Ordered {
            return m.currentLists[i].Number + 1
        }
        if m.currentLists[i].Indent < indent {
            return 1
        }
    }
    return 1
}

func (m *Macchiato) renderLists() string {
    var html strings.Builder
    var openLists []ListType
    var indents []int

    for _, item := range m.currentLists {
        // Close lists if needed
        for len(openLists) > 0 && item.Indent <= indents[len(indents)-1] {
            html.WriteString(m.closeList(openLists[len(openLists)-1]))
            openLists = openLists[:len(openLists)-1]
            indents = indents[:len(indents)-1]
        }

        // Open new list if needed
        if len(openLists) == 0 || item.Indent > indents[len(indents)-1] {
            html.WriteString(m.openList(item.Type))
            openLists = append(openLists, item.Type)
            indents = append(indents, item.Indent)
        }

        html.WriteString(m.renderListItem(item))
    }

    // Close any remaining open lists
    for i := len(openLists) - 1; i >= 0; i-- {
        html.WriteString(m.closeList(openLists[i]))
    }

    m.currentLists = []ListItem{} // Clear the list after rendering
    return html.String()
}

func (m *Macchiato) openList(listType ListType) string {
    switch listType {
    case Ordered:
        return "<ol>\n"
    default:
        return "<ul>\n"
    }
}

func (m *Macchiato) closeList(listType ListType) string {
    switch listType {
    case Ordered:
        return "</ol>\n"
    default:
        return "</ul>\n"
    }
}

func (m *Macchiato) renderListItem(item ListItem) string {
    if item.Type == Ordered {
        return fmt.Sprintf("<li value=\"%d\">%s</li>\n", item.Number, m.parseLine(item.Content))
    }
    return fmt.Sprintf("<li>%s</li>\n", m.parseLine(item.Content))
}

func (m *Macchiato) formatCodeBlock(code, language string) string {
    return fmt.Sprintf(`<div class="code-block">
        <button class="copy-button" onclick="copyCode(this)">Copy</button>
        <pre><code class="language-%s">%s</code></pre>
    </div>`, language, html.EscapeString(code))
}

func (m *Macchiato) formatCustomContainer(content, containerType string) string {
    return fmt.Sprintf("<div class=\"custom-container %s\">%s</div>\n", containerType, m.Parse(content))
}

func (m *Macchiato) handleTable(rows []string) string {
    if len(rows) < 2 {
        return "" // Not a valid table
    }

    var tableHTML strings.Builder
    tableHTML.WriteString("<table>\n")

    // Parse header
    headerCells := m.parseTableRow(rows[0])
    tableHTML.WriteString("  <thead>\n    <tr>\n")
    for _, cell := range headerCells {
        tableHTML.WriteString(fmt.Sprintf("      <th>%s</th>\n", cell))
    }
    tableHTML.WriteString("    </tr>\n  </thead>\n")

    // Check if the second row is a separator
    if m.isTableSeparator(rows[1]) {
        // Parse body starting from the third row
        tableHTML.WriteString("  <tbody>\n")
        for _, row := range rows[2:] {
            tableHTML.WriteString(m.formatTableRow(row))
        }
        tableHTML.WriteString("  </tbody>\n")
    } else {
        // No separator, treat all rows as body
        tableHTML.WriteString("  <tbody>\n")
        for _, row := range rows[1:] {
            tableHTML.WriteString(m.formatTableRow(row))
        }
        tableHTML.WriteString("  </tbody>\n")
    }

    tableHTML.WriteString("</table>\n")
    return tableHTML.String()
}

func (m *Macchiato) parseTableRow(row string) []string {
    row = strings.Trim(row, "|")
    cells := strings.Split(row, "|")
    for i, cell := range cells {
        cells[i] = strings.TrimSpace(cell)
    }
    return cells
}

func (m *Macchiato) formatTableRow(row string) string {
    cells := m.parseTableRow(row)
    var rowHTML strings.Builder
    rowHTML.WriteString("    <tr>\n")
    for _, cell := range cells {
        rowHTML.WriteString(fmt.Sprintf("      <td>%s</td>\n", m.parseLine(cell)))
    }
    rowHTML.WriteString("    </tr>\n")
    return rowHTML.String()
}

func (m *Macchiato) isTableSeparator(row string) bool {
    row = strings.Trim(row, "|")
    cells := strings.Split(row, "|")
    for _, cell := range cells {
        trimmed := strings.TrimSpace(cell)
        if !strings.HasPrefix(trimmed, "-") {
            return false
        }
    }
    return true
}

func (m *Macchiato) handleFootnoteReference(groups []string) string {
    ref := groups[1]
    return fmt.Sprintf("<sup><a href=\"#fn-%s\" id=\"fnref-%s\">[%s]</a></sup>", ref, ref, ref)
}

func (m *Macchiato) handleFootnoteDefinition(groups []string) string {
    ref := groups[1]
    content := groups[2]
    m.footnotes[ref] = content
    return "" // Footnote definitions are not rendered inline
}

func (m *Macchiato) renderFootnotes() string {
    if len(m.footnotes) == 0 {
        return ""
    }
    var html strings.Builder
    html.WriteString("<hr>\n<ol class=\"footnotes\">\n")
    for ref, content := range m.footnotes {
        html.WriteString(fmt.Sprintf("<li id=\"fn-%s\">%s <a href=\"#fnref-%s\">↩</a></li>\n", ref, content, ref))
    }
    html.WriteString("</ol>\n")
    return html.String()
}

func loadDataFile(filePath string) (string, error) {
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return "", fmt.Errorf("error resolving path: %v", err)
    }

    content, err := ioutil.ReadFile(absPath)
    if err != nil {
        return "", fmt.Errorf("error reading file: %v", err)
    }

    return string(content), nil
}

func saveHTMLFile(filePath string, content string) error {
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return fmt.Errorf("error resolving absolute path: %v", err)
    }

    err = ioutil.WriteFile(absPath, []byte(content), 0644)
    if err != nil {
        return fmt.Errorf("error writing file: %v", err)
    }

    return nil
}

func main() {

    cssPath := flag.String("style", "static/style.css", "Path to the CSS file (optional)")
    flag.Parse()

    args := flag.Args()
    if len(args) < 1 {
        fmt.Println("Usage: macchiato <input_file> [output_file] [--style=<css_file>]")
        os.Exit(1)
    }

    inputPath := args[0]
    outputPath := ""
    if len(args) > 1 {
        outputPath = args[1]
    }

    markdown, err := loadDataFile(inputPath)
    if err != nil {
        fmt.Printf("Error loading Markdown file: %v\n", err)
        os.Exit(1)
    }

    css, err := loadDataFile(*cssPath)
    if err != nil {
        fmt.Printf("Warning: CSS file '%s' not found. Using default CSS.\n", *cssPath)
        css, err = loadDataFile("static/style.css")
        if err != nil {
            fmt.Printf("Error: Default CSS file not found.\n")
            os.Exit(1)
        }
    }

    htmlTemplatePath := "static/main.html"
    htmlTemplate, err := loadDataFile(htmlTemplatePath)
    if err != nil {
        fmt.Printf("Error loading HTML template: %v\n", err)
        os.Exit(1)
    }

    parser := NewMacchiato()
    parsedContent := parser.Parse(markdown)

    fullHTML := fmt.Sprintf(htmlTemplate, filepath.Base(inputPath), css, parsedContent)

    if outputPath != "" {
        err = saveHTMLFile(outputPath, fullHTML)
        if err != nil {
            fmt.Printf("Error saving file: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("HTML output saved to %s\n", outputPath)
    } else {
        fmt.Println(fullHTML)
    }
}
