# LUCID Paper - VLDB Demo Track
.PHONY: all clean watch

MAIN = main

all: $(MAIN).pdf

$(MAIN).pdf: $(MAIN).tex sections/*.tex references.bib
	pdflatex -interaction=nonstopmode $(MAIN).tex
	bibtex $(MAIN)
	pdflatex -interaction=nonstopmode $(MAIN).tex
	pdflatex -interaction=nonstopmode $(MAIN).tex

watch:
	latexmk -pdf -pvc -interaction=nonstopmode $(MAIN).tex

draft:
	pdflatex -interaction=nonstopmode $(MAIN).tex

clean:
	rm -f *.aux *.log *.out *.toc *.bbl *.blg *.fdb_latexmk *.fls *.synctex.gz *.run.xml
	rm -f $(MAIN).pdf
