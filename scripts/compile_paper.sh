#!/bin/bash
# ATLAS (VLDB Demo Track) paper compilation script
# Uses pdflatex + bibtex (standard LaTeX for English papers)

set -e

# Prefer TeX Live 2025 if available
if [[ -d "/usr/local/texlive/2025/bin/x86_64-linux" ]]; then
  export PATH="/usr/local/texlive/2025/bin/x86_64-linux:$PATH"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PAPER_DIR="$ROOT_DIR/paper"
OUTPUT_DIR="$PAPER_DIR"

CLEAN=0
if [[ "${1:-}" == "--clean" ]]; then
  CLEAN=1
fi

if [[ ! -d "$PAPER_DIR" ]]; then
  echo "ERROR: paper directory not found: $PAPER_DIR" >&2
  exit 1
fi

cd "$PAPER_DIR"

if [[ $CLEAN -eq 1 ]]; then
  echo "[clean] removing build artifacts..."
  rm -f \
    main.aux main.bbl main.blg main.log main.out main.toc main.lof main.lot \
    main.fdb_latexmk main.fls main.synctex.gz main.synctex\(busy\) \
    main.run.xml main.pdf
  echo "[clean] done."
  if [[ "${2:-}" == "" ]]; then
    exit 0
  fi
fi

echo "=========================================="
echo "  Compiling ATLAS — VLDB Demo Track Paper"
echo "  Directory: $PAPER_DIR"
echo "=========================================="

echo ""
echo "[1/4] pdflatex (pass 1)..."
pdflatex -interaction=nonstopmode -file-line-error main.tex

echo ""
echo "[2/4] bibtex..."
bibtex main || true

echo ""
echo "[3/4] pdflatex (pass 2)..."
pdflatex -interaction=nonstopmode -file-line-error main.tex

echo ""
echo "[4/4] pdflatex (pass 3)..."
pdflatex -interaction=nonstopmode -file-line-error main.tex

echo ""
echo "=========================================="
echo "  ✓ Compilation complete!"
echo "=========================================="
echo "  PDF: $OUTPUT_DIR/main.pdf"
ls -lh "$OUTPUT_DIR/main.pdf"
echo ""