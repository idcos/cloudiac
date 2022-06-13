package services

import (
	"testing"
)

func TestCheckPasswordFormat(t *testing.T) {
	pwCase := []string{
		"2fuDdz",
		"&IrE#C",
		"%MVi34",
		"e5Tuxq",
		"W&q%8d",
		"T1H3Wu",
		"dwU%uh",
		"BDe3tHbOJDTea5r",
		"@X*AI2nHo#RtJl9",
		"MOzYvdGLydZL#@M",
		"U2nuU678xLygw#k",
		"^t0$tbnVuT3!nG8",
		"1hIFw#qfZuM4TIs8UvTz",
		"HDc2CKi$Y6vew!H$d0tX",
		"*oZKbqdpLnEmOK92ycPn",
		"ytU42*K6o2AK%2%mQId$",
		"2Pwi$byqA1be2z1ghLXA",
		"cyh1T$zo^up*X3hG4%4UCpR$v",
		"I&6Z#dLpub0Ph#UQfdyFT!2Z$",
		"p5xOSukHJjSGFFv*CPLT1UvHd",
		"qN5XMTmyl9!17W1r3dWsB!ZoN",
		"izZz2pdXI#jgWT0Mn5WpwgtXj",
		"@BPDCa0EZusDTzMUxzHab&PdEBPU*t",
		"RxlMWRhvKp#&luau19sH@awIRV$gDX",
		"@iLlf0jZ$2r0Y@YHq!FMR7@DX5on2Y",
		"4uoH5MngG6mZ$5GfU3*p&Z%fgbM@u1",
		"dLUqPOXvPYRTGV7Ydq@LXTs@5sf5!N",
		"xG79HHqhJ*EQknQfXu9bNCy6JdaOT0",
		"!vpJwv3Ie6L3p!vNj&cikb%3QN^Us3",
		"GU7YLucIdg7&4Ru%^oGgC8UF1XS9JQ",
		"Bm3!3iJZL&GdWFGQACxcsJbpJgdqnh",
		"XL15TkbeBMgW*!3Z%$idCXCu30xwF$",
	}
	for _, c := range pwCase {
		if err := CheckPasswordFormat(c); err != nil {
			t.Errorf("password %s, err %v", c, err)
		}
	}
}
