#define _POSIX_C_SOURCE 200809L

#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include <fcntl.h>
#include <unistd.h>
#include <termios.h>

#include <err.h>

#define MIN(a, b) ((a) < (b) ? (a) : (b))
#define MAX(a, b) ((a) > (b) ? (a) : (b))
#define CTRL(k) ((k) & 0x1f)

char	*ttbuf;
size_t	 ttbufsz;
size_t	 ttbufpos;

int	srows;
int	scols;

struct termios	ios;
int		tty;

static void	display(void);
static void	ttmove(int, int);
static void	tteeol(void);
static void	tteeop(void);
static void	ttprintf(const char *, ...);
static void	ttputs(const char *);
static void	ttputc(char);
static void	ttflush(void);
static void	fini(void);
static int	setios(int);

void
display(void)
{
	tteeop();
	ttmove(0, 0);
	ttflush();
}

void
ttmove(int row, int col)
{
	ttprintf("\x1b[%d;%dH", row + 1, col + 1);
}

void
tteeol(void)
{
	ttputs("\x1b[2K");
}

void
tteeop(void)
{
	ttputs("\x1b[2J");
}

void
ttprintf(const char *fmt, ...)
{
	va_list ap;
	char s[1024];
	int sz;

	va_start(ap, fmt);
	sz = vsprintf(s, fmt, ap);
	va_end(ap);
	if (ttbufpos + sz > ttbufsz)
		ttflush();
	memcpy(ttbuf + ttbufpos, s, sz);
	ttbufpos += sz;
}

void
ttputs(const char *s)
{
	size_t sz;
	
	sz = strlen(s);
	if (ttbufpos + sz > ttbufsz)
		ttflush();
	memcpy(ttbuf + ttbufpos, s, sz);
	ttbufpos += sz;
}

void
ttputc(char c)
{
	if (ttbufpos >= ttbufsz) ttflush();
	ttbuf[ttbufpos++] = c;
}

void
ttflush(void)
{
	write(1, ttbuf, MIN(ttbufpos, ttbufsz));
	ttbufpos = 0;
}

void
fini(void)
{
	setios(0);
}

int
setios(int raw)
{
	struct termios *p = &ios, rawios;

	if (raw) {
		if (tcgetattr(tty, &ios) == -1) return -1;
		rawios = ios;
		rawios.c_lflag &= ~(ICANON | ECHO);
		rawios.c_iflag &= ~(IXON);
		p = &rawios;
	}
	if (tcsetattr(tty, TCSANOW, p) == -1) return -1;
	return 0;
}

int
main(void)
{
	int i = 0;
	char *ssrows, *sscols;

	errno = 0;

	tty = open("/dev/tty", O_RDWR);
	if (tty == -1)
		err(1, "/dev/tty");
	if (setios(1) == -1)
		err(1, "can't set tty settings");
	atexit(fini);

	srows = strtol((ssrows = getenv("LINES")) ? ssrows : "", NULL, 10);
	scols = strtol((sscols = getenv("COLUMNS")) ? sscols : "", NULL, 10);
	if (!srows || !scols) {
		srows = 24;
		scols = 80;
	}
	ttbufsz = (srows * scols) * 2;
	ttbuf = malloc(ttbufsz);
	if (!ttbuf)
		err(1, NULL);

	for (;;) {
		char c;
		display();
		if (read(tty, &c, 1) == -1)
			err(1, "read key");
		switch (c) {
		case CTRL('q'): goto end;
		}
	}
end:

	return 0;
}
