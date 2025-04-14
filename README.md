# qrpdfstream

A high-performance streaming PDF generator for embedding large sets of QR codes. Memory-efficient, customizable layout, and fully compatible with PDF viewers.

## Features

- Dynamic QR code layout based on size
- Efficient memory usage via streaming writes
- Custom headers, footers, and logos
- Built-in parallel QR code generation
- Fully compatible with macOS Preview, Acrobat Reader, etc.

## Project Structure

qrpdfstream/
├── go.mod
├── LICENSE
├── README.md
├── cmd/
│   └── generate/              # CLI or example app
│       └── main.go
├── internal/
│   └── utils.go               # helper functions (e.g. escape string, image utils)
├── pdf/
│   └── writer.go              # PDF struct and streaming logic
├── qrcode/
│   └── generator.go           # QR code generation logic
├── layout/
│   └── layout.go              # Layout calculator
└── assets/
    └── qrcite.png            # Sample logo (for example/demo)

## Install

```bash
go get github.com/yourusername/qrpdfstream
