package main

import "C"
import (
	"fmt"
	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	"os"
)

type StreamContext struct {
	decCtx []*avcodec.Context
	encCtx []*avcodec.Context
}

var streamCtx = &StreamContext{

}

var ifmtCtx1 avformat.Context
var ofmtCtx *avformat.Context

func openInputFile(filename string) (ret int) {
	ifmtCtx := avformat.AvformatAllocContext()

	if ret = avformat.AvformatOpenInput(&ifmtCtx, filename, nil, nil); ret < 0 {
		fmt.Errorf("%s\n", "打开Input文件众所周知失败")
		return
	}

	// 找到信息流信息
	if ret = ifmtCtx.AvformatFindStreamInfo(nil); ret < 0 {
		fmt.Errorf("找不到流的信息\n")
		return
	}

	streamCount := int(ifmtCtx.NbStreams())
	streamCtx.decCtx = make([]*avcodec.Context, streamCount)
	for i := 0; i < streamCount; i++ {

		stream := ifmtCtx.Streams()[i]
		// 解码器 Decoder
		dec := avcodec.AvcodecFindDecoder(stream.CodecParameters().AvCodecGetId())
		if dec == nil {
			fmt.Errorf("Failed to find decoder for stream #%d\n", i)
			// AVERROR_DECODER_NOT_FOUND
			return
		}
		// 解码器上下文
		codecCtx := dec.AvcodecAllocContext3()
		if codecCtx == nil {
			fmt.Errorf("Failed to allocate the decoder context for stream #%d\n", i)
			return
		}

		// TODO avcodec_parameters_to_context

		if codecCtx.CodecType() == avformat.AVMEDIA_TYPE_VIDEO ||
			codecCtx.CodecType() == avformat.AVMEDIA_TYPE_AUDIO {

			if codecCtx.CodecType() == avformat.AVMEDIA_TYPE_VIDEO {
				//frameRate := ifmtCtx.AvGuessFrameRate(stream, nil)
				// fmt.Print(frameRate.Num())
				//TODO codec_ctx->framerate = av_guess_frame_rate(ifmt_ctx, stream, NULL);
			}

			if ret = codecCtx.AvcodecOpen2(dec, nil); ret < 0 {
				fmt.Errorf("Failed to open decoder for stream #%d\n", i)
				return
			}
		}
		streamCtx.decCtx[i] = codecCtx
		//fmt.Printf("stream #%d. AvCodecGetId: %v. AvCodecGetType: %v\n", i, stream.CodecParameters().AvCodecGetId(), stream.CodecParameters().AvCodecGetType())
	}
	// 输出格式相关
	fmt.Println("----- input file info -----")
	ifmtCtx.AvDumpFormat(0, filename, 0)
	fmt.Println("----- input file info -----")
	return
}

func openOutputFile(filename string) (ret int) {

	var encCtx *avcodec.Context

	ofmtCtx := avformat.AvformatAllocContext()
	avformat.AvformatAllocOutputContext2(&ofmtCtx, nil, "mp4", filename)
	if ofmtCtx == nil {
		// AVERROR_UNKNOWN
		fmt.Errorf("Could not create output context\n")
		return -1
	}

	streamCount := int(ifmtCtx1.NbStreams())

	for i := 0; i < streamCount; i++ {
		outStream := ofmtCtx.AvformatNewStream(nil)
		if outStream == nil {
			fmt.Errorf("Failed allocating output stream\n")
			// AVERROR_UNKNOWN
			return -1
		}

		//inStream := ifmtCtx.Streams()[i]

		decCtx := streamCtx.decCtx[i]

		if decCtx.CodecType() == avformat.AVMEDIA_TYPE_VIDEO ||
			decCtx.CodecType() == avformat.AVMEDIA_TYPE_AUDIO {

			encoder := avcodec.AvcodecFindEncoder(decCtx.CodecId())
			if encoder == nil {
				fmt.Errorf("Necessary encoder not found\n")
				// AVERROR_INVALIDDATA
				return -1
			}

			encCtx = encoder.AvcodecAllocContext3()
			if encCtx == nil {
				fmt.Errorf("Failed to allocate the encoder context\n")
				return -1
			}

			// copy context
			if encCtx.AvcodecCopyContext(decCtx) != 0 {
				fmt.Errorf("Couldn't copy codec context\n")
				return
			}

			if ret = encCtx.AvcodecOpen2(encoder, nil); ret < 0 {
				fmt.Errorf("Cannot open video encoder for stream #%d\n", i)
				return
			}

			// avcodec_parameters_from_context
			timebase := outStream.TimeBase()
			encCtx.SetTimebase(timebase.Num(), timebase.Num())

			streamCtx.encCtx[i] = encCtx
		} else if decCtx.CodecType() == avformat.AVMEDIA_TYPE_UNKNOWN {
			fmt.Errorf("Elementary stream #%d is of unknown type, cannot proceed\n", i)
			// AVERROR_INVALIDDATA
			return -1
		} else {
			// set timebase

		}
	}
	fmt.Println("----- output file info -----")
	ofmtCtx.AvDumpFormat(0, filename, 1)
	fmt.Println("----- output file info -----")

	return
}

func main() {

	if ret := openInputFile(`/home/test/Desktop/jn.mp4`); ret < 0 {
		os.Exit(ret)
	}
	openOutputFile(`/home/test/Desktop/jn2.mp4`)
}
