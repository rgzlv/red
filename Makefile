san = -fsanitize=undefined,address,leak
dbg = -g $(san)
CPPFLAGS = -std=c89 -Wpedantic
CFLAGS = $(dbg)

red: red.c

clean:
	rm -f red
