package main

import (
	"fmt"
	"os/exec"
	"strings"
)

var flags = [][]string{
	nil,
	{"-empty_hdlr_name", "1"},
	{"-frag_interleave", "999"},
	{"-fragment_index", "2"},
	{"-iods_audio_profile", "0"},
	{"-iods_audio_profile", "1"},
	{"-iods_video_profile", "0"},
	{"-iods_video_profile", "1"},
	{"-ism_lookahead", "1"},
	{"-mov_gamma", "1"},
	{"-movflags", "cmaf"},
	{"-movflags", "dash"},
	{"-movflags", "default_base_moof"},
	{"-movflags", "delay_moov"},
	{"-movflags", "disable_chpl"},
	{"-movflags", "empty_moov"},
	{"-movflags", "faststart"},
	{"-movflags", "frag_custom"},
	{"-movflags", "frag_discont"},
	{"-movflags", "frag_every_frame"},
	{"-movflags", "frag_keyframe"},
	{"-movflags", "global_sidx"},
	{"-movflags", "isml"},
	{"-movflags", "negative_cts_offsets"},
	{"-movflags", "omit_tfhd_offset"},
	{"-movflags", "prefer_icc"},
	{"-movflags", "rtphint"},
	{"-movflags", "separate_moof"},
	{"-movflags", "skip_sidx"},
	{"-movflags", "skip_trailer"},
	{"-movflags", "use_metadata_tags"},
	{"-movflags", "write_colr"},
	{"-movflags", "write_gama"},
	{"-movie_timescale", "1"},
	{"-skip_iods", "0"},
	{"-use_editlist", "0"},
	{"-use_editlist", "1"},
	{"-use_stream_ids_as_track_ids", "1"},
	{"-video_track_timescale", "1"},
	{"-write_btrt", "0"},
	{"-write_btrt", "1"},
	{"-write_prft", "pts"},
	{"-write_prft", "wallclock"},
	{"-write_tmcd", "0"},
	{"-write_tmcd", "1"},
}

func main() {
	for _, flag := range flags {
		arg := []string{
			"-i", "in.mp4",
			"-c", "copy",
			"-frag_size", "9K",
		}
		arg = append(arg, flag...)
		arg = append(arg, "--", strings.Join(flag, ",")+".mp4")
		cmd := exec.Command("ffmpeg", arg...)
		fmt.Println(cmd.Args)
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}
