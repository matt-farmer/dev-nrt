package reports

import (
	"encoding/csv"
	"fmt"

	"github.com/nsip/dev-nrt/records"
	"github.com/tidwall/sjson"
)

type IsrPrinting struct {
	baseReport // embed common setup capability
}

//
// Individual Student Report (ISR) score summaries used by jurisdicitons to
// produce printed reports - very simple summary
//
func IsrPrintingReport() *IsrPrinting {

	r := IsrPrinting{}
	r.initialise("./config/IsrPrinting.toml")
	r.printStatus()

	return &r

}

//
// implement the ...Pipe interface, core work of the
// report engine.
//
func (r *IsrPrinting) ProcessStudentRecords(in chan *records.StudentOrientedRecord) chan *records.StudentOrientedRecord {

	out := make(chan *records.StudentOrientedRecord)
	go func() {
		defer close(out)
		// open the csv file writer, and set the header
		w := csv.NewWriter(r.outF)
		defer r.outF.Close()
		w.Write(r.config.header)
		defer w.Flush()

		for sor := range in {
			if !r.config.activated { // only process if activated
				out <- sor
				continue
			}

			//
			// generate any calculated fields required
			//
			sor.CalculatedFields = r.calculateFields(sor)

			//
			// now loop through the ouput definitions to create a
			// row of results
			//
			var result string
			var row []string = make([]string, 0, len(r.config.queries))
			for _, query := range r.config.queries {
				result = sor.GetValueString(query)
				row = append(row, result)
			}
			// write the row to the output file
			if err := w.Write(row); err != nil {
				fmt.Println("Warning: error writing record to csv:", r.config.name, err)
			}

			out <- sor
		}
	}()
	return out
}

//
// generates a block of json that can be added to the
// record containing values that are not in the original data
//
//
func (r *IsrPrinting) calculateFields(sor *records.StudentOrientedRecord) []byte {

	json := sor.CalculatedFields // maintain exsting calc fields

	//
	// the ISR has slots for a comment, but these are not supplied by the RRD
	// so we add a generic placeholder, the columns can't be removed as they are input to
	// downstream printing systems
	//
	json, _ = sjson.SetBytes(json, "CalculatedFields.ISRCommmentPlaceholder", "")

	return json

}
