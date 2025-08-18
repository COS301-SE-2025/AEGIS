package report

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"math"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf" // or your gofpdf import path
	"golang.org/x/net/html"
)

type PDFRenderOptions struct {
	PageSize        string  // "A4"
	MarginMm        float64 // 15
	MaxImageWidthMm float64 // 180
	FontRegular     string  // "NotoSans-Regular.ttf"
	FontBold        string  // "NotoSans-Bold.ttf"
	FontItalic      string  // "NotoSans-Italic.ttf"
	FontBoldItalic  string  // "NotoSans-BoldItalic.ttf"
	BaseFontFamily  string  // "NotoSans"
	HeadingColorRGB [3]int  // e.g., dark gray
	TextColorRGB    [3]int
	TableHeaderRGB  [3]int
	BorderGrayRGB   [3]int
}

func NewPDF(opts PDFRenderOptions) *gofpdf.Fpdf {
	ps := opts.PageSize
	if ps == "" {
		ps = "A4"
	}
	pdf := gofpdf.New("P", "mm", ps, "")
	pdf.SetMargins(opts.MarginMm, opts.MarginMm, opts.MarginMm)
	pdf.SetAutoPageBreak(true, opts.MarginMm)
	pdf.AddPage()

	// Register UTF-8 fonts
	dir := func(f string) string { return filepath.Clean(f) }
	pdf.AddUTF8Font(opts.BaseFontFamily, "", dir(opts.FontRegular))
	pdf.AddUTF8Font(opts.BaseFontFamily, "B", dir(opts.FontBold))
	pdf.AddUTF8Font(opts.BaseFontFamily, "I", dir(opts.FontItalic))
	pdf.AddUTF8Font(opts.BaseFontFamily, "BI", dir(opts.FontBoldItalic))

	pdf.SetFont(opts.BaseFontFamily, "", 11)
	pdf.SetTextColor(opts.TextColorRGB[0], opts.TextColorRGB[1], opts.TextColorRGB[2])
	return pdf
}

// ---------- High-level entry points ----------

func RenderReportWithContentGofpdf(pdf *gofpdf.Fpdf, opts PDFRenderOptions, rpt *ReportWithContent) error {
	// Title
	pdf.SetFont(opts.BaseFontFamily, "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Report: %s", strings.TrimSpace(rpt.Metadata.Name)))
	pdf.Ln(12)

	for _, sec := range rpt.Content {
		// Section heading
		if t := strings.TrimSpace(sec.Title); t != "" {
			pdf.SetFont(opts.BaseFontFamily, "B", 13)
			pdf.Cell(0, 8, t)
			pdf.Ln(8)
		}
		pdf.SetFont(opts.BaseFontFamily, "", 11)

		htmlStr := strings.TrimSpace(sec.Content)
		if htmlStr == "" || htmlStr == "<p></p>" || htmlStr == "<p><br></p>" {
			pdf.MultiCell(0, 6, "(No content provided)", "", "", false)
			pdf.Ln(3)
			continue
		}

		if err := renderHTMLFragment(pdf, opts, htmlStr); err != nil {
			// Fallback: dump as plain text if parsing fails
			pdf.MultiCell(0, 6, stripTags(htmlStr), "", "", false)
		}
		pdf.Ln(3)

		// Add a new page if we’re close to the bottom of the printable area
		_, pageH := pdf.GetPageSize()
		_, _, _, bottom := pdf.GetMargins()
		if pdf.GetY() > pageH-bottom-5 { // keep a small safety gap (5 mm)
			pdf.AddPage()
		}

	}
	return nil
}

// ---------- HTML parsing & rendering ----------

type listContext struct {
	Ordered bool
	Index   int
	Indent  int
}

func renderHTMLFragment(pdf *gofpdf.Fpdf, opts PDFRenderOptions, htmlStr string) error {
	nd, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return err
	}

	// default style
	ts := textStyle{
		B: false, I: false, U: false,
		Color: opts.TextColorRGB,
		Size:  11,
	}
	listStack := []listContext{}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			switch strings.ToLower(n.Data) {
			case "h1", "h2", "h3":
				pdf.Ln(2)
				size := map[string]float64{"h1": 18, "h2": 15, "h3": 13}[strings.ToLower(n.Data)]
				applyTextStyle(pdf, opts, textStyle{B: true, Size: size, Color: opts.HeadingColorRGB})
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				pdf.Ln(4)
				restoreTextStyle(pdf, opts, &ts)
			case "p", "div":
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				pdf.Ln(6)
			case "br":
				pdf.Ln(5)
			case "strong", "b":
				old := ts
				ts.B = true
				applyTextStyle(pdf, opts, ts)
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				ts = old
				applyTextStyle(pdf, opts, ts)
			case "em", "i":
				old := ts
				ts.I = true
				applyTextStyle(pdf, opts, ts)
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				ts = old
				applyTextStyle(pdf, opts, ts)
			case "u":
				old := ts
				ts.U = true
				applyTextStyle(pdf, opts, ts)
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				ts = old
				applyTextStyle(pdf, opts, ts)
			case "span":
				old := ts
				if col, ok := extractColor(n); ok {
					ts.Color = col
					pdf.SetTextColor(col[0], col[1], col[2])
				}
				if sz, ok := extractFontSize(n); ok {
					ts.Size = sz
					pdf.SetFont(opts.BaseFontFamily, fontStyle(ts), sz)
				}
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				ts = old
				applyTextStyle(pdf, opts, ts)
			case "ul":
				listStack = append(listStack, listContext{Ordered: false, Index: 0, Indent: len(listStack)})
				renderChildrenBlock(pdf, opts, n, &ts, &listStack)
				listStack = listStack[:len(listStack)-1]
			case "ol":
				listStack = append(listStack, listContext{Ordered: true, Index: 0, Indent: len(listStack)})
				renderChildrenBlock(pdf, opts, n, &ts, &listStack)
				listStack = listStack[:len(listStack)-1]
			case "li":
				if len(listStack) == 0 { // treat as paragraph
					renderChildrenInline(pdf, opts, n, &ts, &listStack)
					pdf.Ln(5)
					break
				}
				ctx := &listStack[len(listStack)-1]
				ctx.Index++

				left, _, right, _ := pdf.GetMargins()
				indentMm := 6.0 * float64(ctx.Indent+1)
				pdf.SetLeftMargin(left + indentMm)
				pdf.SetRightMargin(right)

				label := "•"
				if ctx.Ordered {
					label = fmt.Sprintf("%d.", ctx.Index)
				}
				applyTextStyle(pdf, opts, ts)
				pdf.CellFormat(6, 5, label, "", 0, "", false, 0, "")
				x := pdf.GetX()
				y := pdf.GetY()
				renderChildrenInline(pdf, opts, n, &ts, &listStack)
				// advance to next line after content
				pdf.SetXY(x, y)
				pdf.Ln(6)

				// restore margins
				pdf.SetLeftMargin(left)
				pdf.SetRightMargin(right)
			case "img":
				renderImageNode(pdf, opts, n)
			case "table":
				renderTable(pdf, opts, n, &ts)
			case "blockquote":
				old := ts
				pdf.SetDrawColor(180, 180, 180)
				x := pdf.GetX()
				y := pdf.GetY()
				pdf.Rect(x, y, 2, 6, "F") // simple left bar
				pdf.SetX(x + 4)
				ts.Color = [3]int{100, 100, 100}
				applyTextStyle(pdf, opts, ts)
				renderChildrenBlock(pdf, opts, n, &ts, &listStack)
				ts = old
				applyTextStyle(pdf, opts, ts)
				pdf.Ln(3)
			default:
				// default: render children
				renderChildrenBlock(pdf, opts, n, &ts, &listStack)
			}
		case html.TextNode:
			txt := normalizeSpaces(n.Data)
			if txt != "" {
				pdf.Write(5, txt)
			}
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	// start at body’s children if present
	body := findBody(nd)
	if body == nil {
		walk(nd)
	} else {
		walk(body)
	}
	return nil
}

func renderChildrenInline(pdf *gofpdf.Fpdf, opts PDFRenderOptions, n *html.Node, ts *textStyle, ls *[]listContext) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && strings.ToLower(c.Data) == "p" {
			// inline context: treat nested <p> as plain
			renderChildrenInline(pdf, opts, c, ts, ls)
			pdf.Ln(6)
		} else {
			renderNodeInline(pdf, opts, c, ts, ls)
		}
	}
}
func renderNodeInline(pdf *gofpdf.Fpdf, opts PDFRenderOptions, n *html.Node, ts *textStyle, ls *[]listContext) {
	switch n.Type {
	case html.TextNode:
		txt := normalizeSpaces(n.Data)
		if txt != "" {
			pdf.Write(5, txt)
		}
	case html.ElementNode:
		switch strings.ToLower(n.Data) {
		case "b", "strong":
			old := *ts
			ts.B = true
			applyTextStyle(pdf, opts, *ts)
			renderChildrenInline(pdf, opts, n, ts, ls)
			*ts = old
			applyTextStyle(pdf, opts, *ts)
		case "i", "em":
			old := *ts
			ts.I = true
			applyTextStyle(pdf, opts, *ts)
			renderChildrenInline(pdf, opts, n, ts, ls)
			*ts = old
			applyTextStyle(pdf, opts, *ts)
		case "u":
			old := *ts
			ts.U = true
			applyTextStyle(pdf, opts, *ts)
			renderChildrenInline(pdf, opts, n, ts, ls)
			*ts = old
			applyTextStyle(pdf, opts, *ts)
		case "span":
			old := *ts
			if col, ok := extractColor(n); ok {
				ts.Color = col
				pdf.SetTextColor(col[0], col[1], col[2])
			}
			if sz, ok := extractFontSize(n); ok {
				ts.Size = sz
				pdf.SetFont(opts.BaseFontFamily, fontStyle(*ts), sz)
			}
			renderChildrenInline(pdf, opts, n, ts, ls)
			*ts = old
			applyTextStyle(pdf, opts, *ts)
		case "br":
			pdf.Ln(5)
		case "img":
			renderImageNode(pdf, opts, n)
		default:
			renderChildrenInline(pdf, opts, n, ts, ls)
		}
	default:
		// ignore
	}
}

func renderChildrenBlock(pdf *gofpdf.Fpdf, opts PDFRenderOptions, n *html.Node, ts *textStyle, ls *[]listContext) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderHTMLFragment(pdf, opts, renderOuterHTML(c))
	}
}

// ---------- Tables (basic) ----------

func renderTable(pdf *gofpdf.Fpdf, opts PDFRenderOptions, table *html.Node, ts *textStyle) {
	// Collect rows
	var rows [][]*html.Node
	for tr := table.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type == html.ElementNode && strings.ToLower(tr.Data) == "tr" {
			var cells []*html.Node
			for td := tr.FirstChild; td != nil; td = td.NextSibling {
				if td.Type == html.ElementNode {
					tag := strings.ToLower(td.Data)
					if tag == "td" || tag == "th" {
						cells = append(cells, td)
					}
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		}
	}
	if len(rows) == 0 {
		return
	}

	// Equal widths
	pageW, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usable := pageW - left - right
	cols := len(rows[0])
	colW := usable / float64(cols)

	// Draw rows
	for r, cells := range rows {
		isHeader := r == 0 && anyHeader(cells)
		for c, cell := range cells {
			x := pdf.GetX()
			y := pdf.GetY()

			// Background for header
			if isHeader {
				pdf.SetFillColor(opts.TableHeaderRGB[0], opts.TableHeaderRGB[1], opts.TableHeaderRGB[2])
				pdf.Rect(x, y, colW, 8, "F")
				applyTextStyle(pdf, opts, textStyle{B: true, Size: 11, Color: opts.TextColorRGB})
			} else {
				applyTextStyle(pdf, opts, *ts)
			}

			// Cell text (simplified: gather plain text)
			txt := stripTags(renderInnerText(cell))
			pdf.MultiCell(colW, 6, txt, "1", "L", false)

			// Move to next cell
			if c < cols-1 {
				pdf.SetXY(x+colW, y)
			}
		}
		// New line for next row
		pdf.Ln(-1)
	}
}

func anyHeader(cells []*html.Node) bool {
	for _, n := range cells {
		if strings.ToLower(n.Data) == "th" {
			return true
		}
	}
	return false
}

// ---------- Images ----------

var dataURLRe = regexp.MustCompile(`^data:(image/(?:png|jpeg|jpg));base64,(.*)$`)

func renderImageNode(pdf *gofpdf.Fpdf, opts PDFRenderOptions, n *html.Node) {
	for _, a := range n.Attr {
		if strings.ToLower(a.Key) == "src" {
			src := strings.TrimSpace(a.Val)
			if m := dataURLRe.FindStringSubmatch(src); len(m) == 3 {
				mime := strings.ToLower(m[1])
				raw := m[2]
				raw = strings.ReplaceAll(raw, " ", "")
				raw = strings.ReplaceAll(raw, "\n", "")
				data, err := base64.StdEncoding.DecodeString(raw)
				if err != nil {
					return
				}

				r := bytes.NewReader(data)
				imgType := strings.ToUpper(strings.TrimPrefix(mime, "image/"))
				if imgType == "JPEG" {
					imgType = "JPG"
				}

				// Optional PNG→JPG if huge
				if imgType == "PNG" && len(data) > 1_000_000 {
					if im, err := png.Decode(bytes.NewReader(data)); err == nil {
						var buf bytes.Buffer
						_ = jpeg.Encode(&buf, im, &jpeg.Options{Quality: 85})
						r = bytes.NewReader(buf.Bytes())
						imgType = "JPG"
					}
				}

				name := fmt.Sprintf("img-%d", pdf.PageNo())
				info := pdf.RegisterImageOptionsReader(name, gofpdf.ImageOptions{
					ImageType: imgType, ReadDpi: true,
				}, r)

				w, h := info.Width(), info.Height()
				maxW := opts.MaxImageWidthMm
				if maxW <= 0 {
					maxW = 180
				}
				if w > maxW {
					scale := maxW / w
					w = maxW
					h *= scale
				}
				pageW, _ := pdf.GetPageSize()
				x := (pageW - w) / 2
				y := pdf.GetY()
				pdf.ImageOptions(name, x, y, w, 0, false, gofpdf.ImageOptions{ImageType: imgType}, 0, "")
				pdf.Ln(h + 4)
			}
			break
		}
	}
}

// ---------- Helpers ----------

// Replace these helpers

type textStyle struct {
	B, I, U bool
	Color   [3]int
	Size    float64
}

func fontStyle(ts textStyle) string {
	style := ""
	if ts.B {
		style += "B"
	}
	if ts.I {
		style += "I"
	}
	if ts.U {
		style += "U"
	} // underline is part of the style string
	return style
}

func applyTextStyle(pdf *gofpdf.Fpdf, opts PDFRenderOptions, ts textStyle) {
	pdf.SetFont(opts.BaseFontFamily, fontStyle(ts), ts.Size)
	pdf.SetTextColor(ts.Color[0], ts.Color[1], ts.Color[2])
}

func restoreTextStyle(pdf *gofpdf.Fpdf, opts PDFRenderOptions, ts *textStyle) {
	applyTextStyle(pdf, opts, *ts)
}

func extractColor(n *html.Node) ([3]int, bool) {
	for _, a := range n.Attr {
		if strings.ToLower(a.Key) == "style" {
			if v, ok := parseStyleColor(a.Val); ok {
				return v, true
			}
		}
		if strings.ToLower(a.Key) == "color" {
			if c, ok := parseHexColor(a.Val); ok {
				return c, true
			}
		}
	}
	return [3]int{}, false
}
func extractFontSize(n *html.Node) (float64, bool) {
	for _, a := range n.Attr {
		if strings.ToLower(a.Key) == "style" {
			if v, ok := parseStyleFontSize(a.Val); ok {
				return v, true
			}
		}
	}
	return 0, false
}
func parseStyleColor(style string) ([3]int, bool) {
	for _, part := range strings.Split(style, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			continue
		}
		if strings.TrimSpace(strings.ToLower(kv[0])) == "color" {
			if c, ok := parseHexColor(strings.TrimSpace(kv[1])); ok {
				return c, true
			}
		}
	}
	return [3]int{}, false
}
func parseStyleFontSize(style string) (float64, bool) {
	for _, part := range strings.Split(style, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) != 2 {
			continue
		}
		if strings.TrimSpace(strings.ToLower(kv[0])) == "font-size" {
			v := strings.TrimSpace(kv[1])
			// support "12pt", "14px" (rough conversion), or raw number as point
			if strings.HasSuffix(v, "pt") {
				if f, err := strconv.ParseFloat(strings.TrimSuffix(v, "pt"), 64); err == nil {
					return f, true
				}
			} else if strings.HasSuffix(v, "px") {
				if f, err := strconv.ParseFloat(strings.TrimSuffix(v, "px"), 64); err == nil {
					return pxToPt(f), true
				}
			} else if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}
func pxToPt(px float64) float64 { return math.Round(px*0.75*100) / 100.0 } // 96dpi approx

func parseHexColor(s string) ([3]int, bool) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if len(s) == 3 { // short #rgb
		r, _ := strconv.ParseInt(strings.Repeat(string(s[0]), 2), 16, 0)
		g, _ := strconv.ParseInt(strings.Repeat(string(s[1]), 2), 16, 0)
		b, _ := strconv.ParseInt(strings.Repeat(string(s[2]), 2), 16, 0)
		return [3]int{int(r), int(g), int(b)}, true
	}
	if len(s) == 6 {
		r, _ := strconv.ParseInt(s[0:2], 16, 0)
		g, _ := strconv.ParseInt(s[2:4], 16, 0)
		b, _ := strconv.ParseInt(s[4:6], 16, 0)
		return [3]int{int(r), int(g), int(b)}, true
	}
	return [3]int{}, false
}

func normalizeSpaces(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(s), " ")
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && strings.ToLower(n.Data) == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if b := findBody(c); b != nil {
			return b
		}
	}
	return nil
}

func renderOuterHTML(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}
func renderInnerText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(renderOuterHTML(c))
	}
	return b.String()
}
func stripTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	// decode common entities
	s, _ = url.PathUnescape(strings.ReplaceAll(s, "&nbsp;", " "))
	return normalizeSpaces(s)
}
