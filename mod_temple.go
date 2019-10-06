package hermes

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/szampardi/xprint/temple"
)

func unsafeMode() bool {
	envvar, err := strconv.ParseBool(os.Getenv("XPRINT_UNSAFE"))
	if err != nil {
		return false
	}
	return envvar
}

func process(d Data, t Templates) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	if len(t) > 0 {
		argTemplates := map[string]string{}
		localTemplates := []string{}
		for n, t := range t {
			if !t.IsFile {
				if len(t.S) > 0 {
					argTemplates[fmt.Sprintf("opt%d", n)] = t.S
				}
			} else {
				localTemplates = append(localTemplates, t.S)
			}
		}
		tpl, tplList, err := temple.FnMap.BuildTemplate(temple.EnableUnsafeFunctions, hex.EncodeToString(temple.Random(12)), "", argTemplates, localTemplates...)
		if err != nil {
			return nil, err
		}

		if len(tplList) > 1 {
			err = tpl.Execute(buf, d)
		} else {
			err = tpl.ExecuteTemplate(buf, tplList[0], d)
		}
		if err != nil {
			return nil, err
		}
		if debug {
			temple.Tracking.Wait()
		}
	} else {
		rawMessage := func() (err error) {
			if !argsfirst {
				_, err = fmt.Fprintf(buf, "%s", d.Stdin)
				if err != nil {
					return err
				}
			} else {
				defer func() error {
					_, err = fmt.Fprintf(buf, "%s", d.Stdin)
					if err != nil {
						return err
					}
					return nil
				}()
			}
			for _, s := range d.Args {
				_, err = fmt.Fprintf(buf, "%s", s)
				if err != nil {
					return err
				}
			}
			return nil
		}
		if err := rawMessage(); err != nil {
			return nil, err
		}
	}
	return buf, nil
}
