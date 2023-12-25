package sofia

import (
   "net/http"
   "testing"
)

func Test_File(t *testing.T) {
   res, err := http.Get("https://redirector.us-east-1.prod-a.boltdns.net/v1/6245817279001/4a947ef9-6981-46a6-916c-27f57bb91326/xdb/69683dd1-74bf-43c2-9888-a1ffb8d67485/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      t.Fatal(res.Status)
   }
   var f File
   if err := f.Decode(res.Body); err != nil {
      t.Fatal(err)
   }
}
