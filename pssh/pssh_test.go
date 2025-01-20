package pssh

import (
   "encoding/base64"
   "encoding/json"
   "os"
   "testing"
)

const cenc_pssh = "AAACJnBzc2gAAAAAmgTweZhAQoarkuZb4IhflQAAAgYGAgAAAQABAPwBPABXAFIATQBIAEUAQQBEAEUAUgAgAHgAbQBsAG4AcwA9ACIAaAB0AHQAcAA6AC8ALwBzAGMAaABlAG0AYQBzAC4AbQBpAGMAcgBvAHMAbwBmAHQALgBjAG8AbQAvAEQAUgBNAC8AMgAwADAANwAvADAAMwAvAFAAbABhAHkAUgBlAGEAZAB5AEgAZQBhAGQAZQByACIAIAB2AGUAcgBzAGkAbwBuAD0AIgA0AC4AMAAuADAALgAwACIAPgA8AEQAQQBUAEEAPgA8AFAAUgBPAFQARQBDAFQASQBOAEYATwA+ADwASwBFAFkATABFAE4APgAxADYAPAAvAEsARQBZAEwARQBOAD4APABBAEwARwBJAEQAPgBBAEUAUwBDAFQAUgA8AC8AQQBMAEcASQBEAD4APAAvAFAAUgBPAFQARQBDAFQASQBOAEYATwA+ADwASwBJAEQAPgBBAEEAQQBBAEEARQBaAG4AMgBVAEIAdQB3AFQAUwBSAEIALwA4AGwANwB3AD0APQA8AC8ASwBJAEQAPgA8AEMASABFAEMASwBTAFUATQA+AHkAUQB2AGMAbABUAGMAdQBrAGoAbwA9ADwALwBDAEgARQBDAEsAUwBVAE0APgA8AC8ARABBAFQAQQA+ADwALwBXAFIATQBIAEUAQQBEAEUAUgA+AA=="

func TestPssh(t *testing.T) {
   data, err := base64.StdEncoding.DecodeString(cenc_pssh)
   if err != nil {
      t.Fatal(err)
   }
   var pssh Box
   n, err := pssh.BoxHeader.Decode(data)
   if err != nil {
      t.Fatal(err)
   }
   err = pssh.Read(data[n:])
   if err != nil {
      t.Fatal(err)
   }
   encode := json.NewEncoder(os.Stdout)
   encode.SetIndent("", " ")
   err = encode.Encode(pssh)
   if err != nil {
      t.Fatal(err)
   }
   message := protobuf.Message{}
   
   message.Unmarsh
   
   
   
   
   
}
