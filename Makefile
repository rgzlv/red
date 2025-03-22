san = -fsanitize=undefined,address,leak
dbg = -g
CPPFLAGS = -std=c89 -Wpedantic
CFLAGS = $(dbg) $(san)

red: red.c

clean:
	rm -f red
