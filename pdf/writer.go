package pdf

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Raj63/qrpdfstream/internal"
	"github.com/Raj63/qrpdfstream/qrcode"
)

const (
	chunkSize      = 5 * 1024 * 1024  // 5 MB
	bufferCapacity = 10 * 1024 * 1024 // 10 MB or whatever you feel is optimal
)

type PDF struct {
	mu      sync.Mutex
	writer  io.Writer
	buffer  *bytes.Buffer
	xref    []int
	objects int
}

func NewPDF(w io.Writer) *PDF {
	// Preallocate buffer with fixed capacity to reuse memory
	buf := bytes.NewBuffer(make([]byte, 0, bufferCapacity))

	pdf := &PDF{
		writer:  w,
		buffer:  buf,
		xref:    []int{0},
		objects: 1,
	}
	// start of pdf declaration
	pdf.buffer.WriteString("%PDF-1.7\n%\xFF\xFF\xFF\xFF\n")
	pdf.writeObject("1 0 obj\n<< >>\nendobj\n") // empty object to reserve slot

	// Built-in font
	pdf.writeObject("2 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	pdf.flushBuffer()
	return pdf
}

// Optimized writeObject
func (pdf *PDF) writeObject(obj string) int {
	offset := pdf.buffer.Len()
	pdf.xref = append(pdf.xref, offset)
	_, _ = pdf.buffer.WriteString(obj)
	pdf.objects++
	pdf.maybeFlush()
	return pdf.objects
}

// Write raw RGB image data to PDF
// Optimized image encoder (raw RGB + compression)
func (pdf *PDF) AddRawImage(img image.Image, width, height int) int {
	rgb := make([]byte, 0, width*height*3) // Preallocate exact size
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgb = append(rgb, byte(r>>8), byte(g>>8), byte(b>>8))
		}
	}

	comp := internal.CompressStream(rgb)

	// Construct image header (note: object ID will be set in critical section)
	header := func(objNum int) string {
		return fmt.Sprintf(
			"%d 0 obj\n<< /Type /XObject /Subtype /Image /Width %d /Height %d "+
				"/ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n",
			objNum, width, height, len(comp),
		)
	}

	return pdf.writeStreamObject(comp, header)
}

// Thread-safe and optimized
func (pdf *PDF) writeStreamObject(data []byte, headerFunc func(objNum int) string) int {
	pdf.mu.Lock()
	defer pdf.mu.Unlock()

	objNum := pdf.objects + 1
	header := headerFunc(objNum)

	offset := pdf.buffer.Len()
	pdf.xref = append(pdf.xref, offset)
	pdf.buffer.WriteString(header)
	pdf.buffer.Write(data)
	pdf.buffer.WriteString("\nendstream\nendobj\n")
	pdf.objects++

	if pdf.buffer.Len() > chunkSize {
		pdf.flushBuffer()
	}

	return objNum
}

// Adds a page with QR images
func (pdf *PDF) AddPageWithImages(imageIDs []int, imgSize, columnCount int, prependContent, appendContent string, extraImageIDs []int) int {
	var sb strings.Builder
	sb.WriteString(prependContent)

	x, y := 50, 650
	col := 0

	for _, objID := range imageIDs {
		sb.WriteString(fmt.Sprintf("q %d 0 0 %d %d %d cm /Im%d Do Q\n", imgSize, imgSize, x, y, objID))
		x += imgSize + 20
		col++
		if col == columnCount {
			col, x = 0, 50
			y -= imgSize + 20
		}
	}

	sb.WriteString(appendContent)
	comp := internal.CompressStream([]byte(sb.String()))

	contentHeader := func(objNum int) string {
		return fmt.Sprintf("%d 0 obj\n<< /Length %d /Filter /FlateDecode >>\nstream\n", pdf.objects+1, len(comp))
	}
	contentObjID := pdf.writeStreamObject(comp, contentHeader)

	allIDs := append(imageIDs, extraImageIDs...)
	res := internal.NewResourceDictionary(allIDs)
	resourceObjID := pdf.writeObject(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", pdf.objects+1, res))

	return pdf.writeObject(fmt.Sprintf(
		"%d 0 obj\n<< /Type /Page /Parent 1 0 R /MediaBox [0 0 595 842] "+
			"/Contents %d 0 R /Resources %d 0 R >>\nendobj\n",
		pdf.objects+1, contentObjID, resourceObjID))
}

// Flush the accumulated buffer to the writer
func (pdf *PDF) flushBuffer() {
	if pdf.buffer.Len() == 0 {
		return
	}
	fmt.Printf("Flushing buffer, size: %d\n", pdf.buffer.Len())
	_, err := pdf.writer.Write(pdf.buffer.Bytes())
	if err != nil {
		log.Fatal("Error writing PDF:", err)
	}
	pdf.buffer.Reset()
}

// Flush logic thresholded by chunk size
func (pdf *PDF) maybeFlush() {
	if pdf.buffer.Len() > chunkSize {
		pdf.flushBuffer()
	}
}

func (pdf *PDF) Generate(pageIDs []int) {
	// Pages object
	kids := ""
	for _, id := range pageIDs {
		kids += fmt.Sprintf("%d 0 R ", id)
	}
	pagesObjID := pdf.writeObject(fmt.Sprintf(
		"%d 0 obj\n<< /Type /Pages /Count %d /Kids [ %s] >>\nendobj\n",
		pdf.objects+1, len(pageIDs), kids,
	))

	// Catalog
	catalogID := pdf.writeObject(fmt.Sprintf(
		"%d 0 obj\n<< /Type /Catalog /Pages %d 0 R >>\nendobj\n",
		pdf.objects+1, pagesObjID,
	))

	// Xref
	startxref := pdf.buffer.Len()
	pdf.buffer.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", pdf.objects+1))
	for _, offset := range pdf.xref {
		pdf.buffer.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// Trailer
	pdf.buffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root %d 0 R >>\n", pdf.objects+1, catalogID))
	pdf.buffer.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", startxref))

	// Flush the final chunk of data to the output stream
	pdf.flushBuffer()

	// Output PDF
	// io.Copy(pdf.writer, pdf.buffer)
}

func (pdf *PDF) GenerateHeaderFooterContent(title, subtitle string, logo image.Image, isHeader bool) (string, int) {
	const (
		logoWidth     = 50
		logoHeight    = 50
		logoX         = 50
		titleFontSize = 16
		subFontSize   = 10
		infoFontSize  = 8
		titleX        = 120
		infoX         = 400
		topY          = 780
		bottomY       = 50
	)

	y := topY
	if !isHeader {
		y = bottomY
	}

	var content strings.Builder
	imageID := 0

	// Add logo
	if logo != nil {
		resized := internal.FlattenToWhite(internal.ResizeImage(logo, logoWidth, logoHeight))
		imageID = pdf.AddRawImage(resized, logoWidth, logoHeight)
		logoY := y - 20
		content.WriteString(fmt.Sprintf(
			"q %d 0 0 %d %d %d cm /Im%d Do Q\n",
			logoWidth, logoHeight, logoX, logoY, imageID,
		))
	}

	// Title
	if title != "" {
		content.WriteString(fmt.Sprintf(
			"BT /F1 %d Tf %d %d Td (%s) Tj ET\n",
			titleFontSize, titleX, y, internal.EscapePDFString(title),
		))
	}

	// Subtitle or datetime
	if isHeader {
		if subtitle != "" {
			content.WriteString(fmt.Sprintf(
				"BT /F1 %d Tf %d %d Td (%s) Tj ET\n",
				subFontSize, titleX, y-20, internal.EscapePDFString(subtitle),
			))
		}
		now := fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2006 15:04"))
		content.WriteString(fmt.Sprintf(
			"BT /F1 %d Tf %d %d Td (%s) Tj ET\n",
			infoFontSize, infoX, y-30, internal.EscapePDFString(now),
		))
	} else if subtitle != "" {
		content.WriteString(fmt.Sprintf(
			"BT /F1 %d Tf %d %d Td (%s) Tj ET\n",
			infoFontSize, infoX, y-30, internal.EscapePDFString(subtitle),
		))
	}

	return content.String(), imageID
}

// LoadImageFromFile reads an image file (PNG or JPEG) and returns an image.Image
func LoadImageFromFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func GenerateQRCodesParallel(pdf *PDF, dataList []string, size int) []int {
	var wg sync.WaitGroup
	imageIDs := make([]int, len(dataList))

	// Limit parallelism to avoid memory overload (e.g., 8 workers)
	sem := make(chan struct{}, 8)

	for i := 0; i < len(dataList); i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			img := qrcode.GenerateQRCodeImage(dataList[i], size)
			imgID := pdf.AddRawImage(img, size, size)
			imageIDs[i] = imgID
		}(i)
	}

	wg.Wait()
	return imageIDs
}
