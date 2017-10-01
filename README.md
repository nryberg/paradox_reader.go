# Paradox Reader

This is a very simple tool used to convert legacy [Paradox](https://en.wikipedia.org/wiki/Paradox_(database)).  There are a number of Paradox file converters already out there - this is simply an attempt to code it in Go.

## Project Status

Currently unloads the file data from a pdx file.  The project is still in heavy alpha stage - in no case should you be using this in production.

It currently exports the raw data as a "output.csv" CSV file by default - next step is to deal with things like number fields.

## Tasks

* Turn this into an honest cli tool
* Give the user the option to export the table layout in a relatively standard format making it easy to understand the underlying data.
* Work with a real Golang programmer to get the syntax more idiomatic and clear up the many sins committed here.
* Create documentation a little more thorough than an overblown README.md.

## Credit where Credit is due  

The library is built on the following resources as guides for pulling data out of these pdx files.  

* [Prestwood Boards](https://www.prestwoodboards.com/ASPSuite/kb/document_view.asp?qid=100060)
* [Exporting Paradox Tables](http://www.wpuniverse.com/vb/archive/index.php/t-6093.html?s=0b1583bd381106fac6c2b627dbd13e21)

Most importantly, this depends on the [Kevin Mitchell "PARADOX 4.x FILE FORMATS"](https://github.com/nryberg/paradox_reader.go/blob/master/Docs/Kevin_Mitchell_Paradox_Format.pdf), dated May 11th, 1996.  This concise document clearly lays out the process for pulling out useful information from a pdx file.  
