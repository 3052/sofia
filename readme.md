# sofia

> And, in the dream, I knew that he was, going on ahead. He was fixin' to make
> a fire somewhere out there in all that dark and cold. And I knew that
> whenever I got there, he'd be there. And then I woke up.
>
> [No Country for Old Men](//youtube.com/watch?v=GH4IhjtaAUQ) (2007)

## features

* **`Remuxer.Initialize`**: Writes a 16-byte `mdat` (Media Data) header directly to the `io.WriteSeeker`.
* **`Remuxer.AddSegment`**: Appends raw media sample payloads directly to the `io.WriteSeeker` file as segments are processed.
* **`Remuxer.Finish`**: Appends the final `moov` metadata box to the end of the `io.WriteSeeker` file, and seeks backward to overwrite the `mdat` placeholder size with the final calculated byte size.
* **`Decrypt`**: Applies an AES-CTR XOR key stream directly onto the data byte slice. If this slice is backed by a memory-mapped file, it modifies the file on disk in-place.
* **`MoovBox.RemovePssh`**: Mutates the in-memory `MoovBox` to strip out all PSSH (Protection System Specific Header) boxes, altering the structure before it is written to a file.
* **`MoovBox.RemoveMvex`**: Mutates the in-memory `MoovBox` to strip out the `mvex` (Movie Extends) boxes, altering the structure before it is written to a file.
* **`TrakBox.RemoveEdts`**: Mutates the in-memory `TrakBox` to strip out the `edts` (Edit List) boxes, altering the structure before it is written to a file.
* **`StsdBox.RemoveSinf`**: Mutates the in-memory `StsdBox` to strip out the `sinf` (Protection Scheme Information) boxes and alters the entry header format, altering the structure before it is written to a file.
* **`MvhdBox.SetDuration`**: Mutates the in-memory `MvhdBox` to update the total duration of the movie, automatically adjusting the version flag if a 64-bit size is required.
* **`MdhdBox.SetDuration`**: Mutates the in-memory `MdhdBox` to update the media duration, automatically adjusting the version flag if a 64-bit size is required.

## prior art

1. https://github.com/mozilla/mp4parse-rust/issues/415
2. https://github.com/Eyevinn/mp4ff/issues/311
3. https://github.com/alfg/mp4-rust/issues/132
4. https://github.com/yapingcat/gomedia/issues/115
5. https://github.com/alfg/mp4/issues/27
6. https://github.com/abema/go-mp4/issues/13
7. https://github.com/garden4hu/fmp4parser-go/issues/4
8. https://github.com/eswarantg/mp4box/issues/3
9. https://github.com/miquels/mp4/issues/2

## standard

ISO/IEC 14496-12:
<https://wikipedia.org/wiki/ISO_base_media_file_format>

ISO/IEC 14496-14:
<https://wikipedia.org/wiki/MP4_file_format>

ISO/IEC 23001-7:
<https://wikipedia.org/wiki/MPEG_Common_Encryption>

## Discord

https://discord.com/invite/rMFzDRQhSx
