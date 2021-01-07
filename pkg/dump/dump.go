package dump

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"gitlab.com/c0b/go-ordered-json"
)

// Pretty dumps the given data to the given stream so that it looks pretty. If the data is a valid
// JSON document then it will be indented before printing it. If the `jq` tools is available in the
// path then it will be used for syntax highlighting.
func Pretty(stream io.Writer, body []byte) error {
	if len(body) == 0 {
		return nil
	}
	data := ordered.NewOrderedMap()
	err := json.Unmarshal(body, data)
	if err != nil {
		return dumpBytes(stream, body)
	}
	if haveJQ() {
		return dumpJQ(stream, body)
	}
	return dumpJSON(stream, data)
}

// Simple functions exactly the same as Pretty except it uses jq's -c option to condense the
// output to a single line, intended to be used with other resources that require single line
// output.
func Simple(stream io.Writer, body []byte) error {
	if len(body) == 0 {
		return nil
	}
	data := ordered.NewOrderedMap()
	err := json.Unmarshal(body, data)
	if err != nil {
		return dumpBytes(stream, body)
	}
	if haveJQ() {
		return dumpCondensedJQ(stream, body)
	}
	return dumpJSON(stream, data)
}

func dumpBytes(stream io.Writer, data []byte) error {
	_, err := stream.Write(data)
	if err != nil {
		return err
	}
	_, err = stream.Write([]byte("\n"))
	return err
}

func dumpJQ(stream io.Writer, data []byte) error {
	// #nosec 204
	jq := exec.Command("jq", ".")
	jq.Stdin = bytes.NewReader(data)
	jq.Stdout = stream
	jq.Stderr = os.Stderr
	return jq.Run()
}

func dumpCondensedJQ(stream io.Writer, data []byte) error {
	// #nosec 204
	jq := exec.Command("jq", "-c", ".")
	jq.Stdin = bytes.NewReader(data)
	jq.Stdout = stream
	jq.Stderr = os.Stderr
	return jq.Run()
}

func dumpJSON(stream io.Writer, data *ordered.OrderedMap) error {
	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func haveJQ() bool {
	_, err := exec.LookPath("jq")
	return err == nil
}
