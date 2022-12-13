package main

import (
	"flag"
	"time"
)

func RegisterFlags(flags *flag.FlagSet) {
	flags.Int("device", -1, "Capture device")
	flags.Int("proc", -1, "Number of processors, use 0 for sequential")
	flags.Duration("length", 5*time.Second, "Sample window length")
}

func FlagProc(flags *flag.FlagSet) int {
	return flags.Lookup("proc").Value.(flag.Getter).Get().(int)
}

func FlagDevice(flags *flag.FlagSet) int {
	return flags.Lookup("device").Value.(flag.Getter).Get().(int)
}

func FlagLength(flags *flag.FlagSet) time.Duration {
	return flags.Lookup("length").Value.(flag.Getter).Get().(time.Duration)
}

/*
fprintf(stderr, "  -mc N,    --max-context N [%-7d] maximum number of text context tokens to store\n", params.max_context);
fprintf(stderr, "  -ml N,    --max-len N     [%-7d] maximum segment length in characters\n",           params.max_len);
fprintf(stderr, "  -wt N,    --word-thold N  [%-7.2f] word timestamp probability threshold\n",         params.word_thold);
fprintf(stderr, "  -su,      --speed-up      [%-7s] speed up audio by x2 (reduced accuracy)\n",        params.speed_up ? "true" : "false");
fprintf(stderr, "  -tr,      --translate     [%-7s] translate from source language to english\n",      params.translate ? "true" : "false");
fprintf(stderr, "  -di,      --diarize       [%-7s] stereo audio diarization\n",                       params.diarize ? "true" : "false");
fprintf(stderr, "  -otxt,    --output-txt    [%-7s] output result in a text file\n",                   params.output_txt ? "true" : "false");
fprintf(stderr, "  -ovtt,    --output-vtt    [%-7s] output result in a vtt file\n",                    params.output_vtt ? "true" : "false");
fprintf(stderr, "  -osrt,    --output-srt    [%-7s] output result in a srt file\n",                    params.output_srt ? "true" : "false");
fprintf(stderr, "  -owts,    --output-words  [%-7s] output script for generating karaoke video\n",     params.output_wts ? "true" : "false");
fprintf(stderr, "  -ps,      --print-special [%-7s] print special tokens\n",                           params.print_special ? "true" : "false");
fprintf(stderr, "  -pc,      --print-colors  [%-7s] print colors\n",                                   params.print_colors ? "true" : "false");
fprintf(stderr, "  -nt,      --no-timestamps [%-7s] do not print timestamps\n",                        params.no_timestamps ? "false" : "true");
fprintf(stderr, "  -l LANG,  --language LANG [%-7s] spoken language\n",                                params.language.c_str());
fprintf(stderr, "  -m FNAME, --model FNAME   [%-7s] model path\n",                                     params.model.c_str());
fprintf(stderr, "  -f FNAME, --file FNAME    [%-7s] input WAV file path\n",                            "");
*/
