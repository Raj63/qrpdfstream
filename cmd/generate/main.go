package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Raj63/qrpdfstream/layout"
	"github.com/Raj63/qrpdfstream/pdf"
)

func main() {
	outFile, err := os.Create("./qrstreamed_1k.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	p := pdf.NewPDF(outFile)
	var pageIDs []int
	// A4 page dimensions in points
	layout := layout.CalculateLayout(80)

	// Use the calculated values
	fmt.Printf("QR Code Size: %d\n", layout.QRCodeSize)
	fmt.Printf("Columns: %d\n", layout.ColumnCount)
	fmt.Printf("Images per page: %d\n", layout.ImagesPerPage)

	measureExecution(func() {
		logoImg, err := pdf.LoadImageFromFile("./../../assets/qrcite.png")
		if err != nil {
			log.Fatal(err)
		}

		// Add header
		headerStr, headerImgID := p.GenerateHeaderFooterContent("QRCite Report", "Smart Labels", logoImg, true)
		footerStr, footerImgID := p.GenerateHeaderFooterContent("", "Thanks!", nil, false)

		qrcodeSize := layout.QRCodeSize       // 80    // qrcode size
		columnCount := layout.ColumnCount     // 5    // number of columns possible for above qrcodeSize
		imagesPerPage := layout.ImagesPerPage // 35               // Number of QR codes per page possible for above qrcodeSize & columnCount
		// now can we make the qrcodeSize dynamic and then calculate the columnCount and imagesPerPage based on it

		totalQRCodes := 100

		dataList := make([]string, totalQRCodes)
		for i := 0; i < totalQRCodes; i++ {
			dataList[i] = fmt.Sprintf("https://qrcite.com?id=%d", i+1)
		}

		imageIDs := pdf.GenerateQRCodesParallel(p, dataList, qrcodeSize)

		// Stream page generation immediately
		currentPageImages := []int{}

		for i, imgID := range imageIDs {
			currentPageImages = append(currentPageImages, imgID)
			if len(currentPageImages) >= imagesPerPage || i == len(imageIDs)-1 {
				pageID := p.AddPageWithImages(currentPageImages, qrcodeSize, columnCount, headerStr, footerStr, []int{headerImgID, footerImgID})
				pageIDs = append(pageIDs, pageID)
				currentPageImages = []int{} // Reset for the next page
			}
		}

		p.Generate(pageIDs)

		fmt.Println("PDF generated ✅")
	})
}

func measureExecution(fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	fmt.Printf("Execution Time: %v\n", duration)
}
