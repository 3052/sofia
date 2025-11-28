package sofia

import (
   "encoding/binary"
   "errors"
)

// --- Shared Utility Helpers ---

func makeBox(typeStr string, payload []byte) []byte {
   size := 8 + len(payload)
   buf := make([]byte, 8)
   binary.BigEndian.PutUint32(buf[0:4], uint32(size))
   copy(buf[4:8], []byte(typeStr))
   return append(buf, payload...)
}

func uint32ToBytes(v uint32) []byte {
   b := make([]byte, 4)
   binary.BigEndian.PutUint32(b, v)
   return b
}

// FindMoofPtr finds the first MoofBox pointer in a slice of generic boxes.
func FindMoofPtr(boxes []Box) *MoofBox {
   for _, box := range boxes {
      if box.Moof != nil {
         return box.Moof
      }
   }
   return nil
}

// FindMdatPtr finds the first MdatBox pointer in a slice of generic boxes.
func FindMdatPtr(boxes []Box) *MdatBox {
   for _, box := range boxes {
      if box.Mdat != nil {
         return box.Mdat
      }
   }
   return nil
}

// filterMvex removes the 'mvex' atom from the MoovBox children.
func filterMvex(moov *MoovBox) {
   var cleanChildren []MoovChild
   for _, child := range moov.Children {
      if len(child.Raw) >= 8 {
         if string(child.Raw[4:8]) == "mvex" {
            continue
         }
      }
      cleanChildren = append(cleanChildren, child)
   }
   moov.Children = cleanChildren
}

// patchDuration updates the duration field in a raw mvhd or mdhd box.
func patchDuration(boxData []byte, newDuration uint64) error {
   if len(boxData) < 32 {
      return errors.New("box too short to patch duration")
   }

   version := boxData[8] // Offset 8 is Version

   if version == 1 {
      const durationOffset = 32
      if len(boxData) < durationOffset+8 {
         return errors.New("box too short for v1 duration")
      }
      binary.BigEndian.PutUint64(boxData[durationOffset:], newDuration)
   } else {
      const durationOffset = 24
      if len(boxData) < durationOffset+4 {
         return errors.New("box too short for v0 duration")
      }
      if newDuration > 0xFFFFFFFF {
         return errors.New("duration overflows 32-bit field (update init segment to use version 1 boxes)")
      }
      binary.BigEndian.PutUint32(boxData[durationOffset:], uint32(newDuration))
   }
   return nil
}
