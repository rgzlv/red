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

char	*ttbp;
size_t	 ttbsz;
size_t	 ttbpos;

int	srows;
int	scols;

struct termios	ios;
int		tty;

static void 	display(void);
static void 	ttmove(int, int);
static void 	tteeol(void);
static void 	tteeop(void);
static void 	tttidy(void);
static void 	ttprintf(const char *, ...);
static void 	ttputs(const char *);
static void 	ttputc(char);
static void 	ttflush(void);
static int 	ttopen(const char *);
static void	ttclose(void);
static int 	ttios(int);
static void 	panic(const char *, ...);
static void 	fini(void);

void
display(void)
{
	tteeop();
	ttmove(1, 1);
	ttflush();
}

void
ttmove(int row, int col)
{
	ttprintf("\x1b[%d;%dH", row, col);
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
tttidy(void)
{
	ttmove(srows, 0);
	tteeol();
	ttflush();
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
	if (ttbpos + sz > ttbsz)
		ttflush();
	memcpy(ttbp + ttbpos, s, sz);
	ttbpos += sz;
}

void
ttputs(const char *s)
{
	size_t sz;
	
	sz = strlen(s);
	if (ttbpos + sz > ttbsz)
		ttflush();
	memcpy(ttbp + ttbpos, s, sz);
	ttbpos += sz;
}

void
ttputc(char c)
{
	if (ttbpos >= ttbsz) ttflush();
	ttbp[ttbpos++] = c;
}

void
ttflush(void)
{
	write(1, ttbp, MIN(ttbpos, ttbsz));
	ttbpos = 0;
}

int
ttopen(const char *path)
{
	char *ssrows, *sscols;

	tty = open(path, O_RDWR);
	if (tty == -1)
		return -1;
	if (ttios(1) == -1)
		return -1;
	srows = strtol((ssrows = getenv("LINES")) ? ssrows : "", NULL, 10);
	scols = strtol((sscols = getenv("COLUMNS")) ? sscols : "", NULL, 10);
	if (!srows || !scols) {
		srows = 24;
		scols = 80;
	}
	ttbsz = (srows * scols) * 2;
	ttbp = malloc(ttbsz);
	if (!ttbp)
		panic(NULL);
	return tty;
}

void
ttclose()
{
	if (ttios(0) == -1)
		panic("ttclose: ttios");
	close(tty);
	tty = 0;
}

int
ttios(int raw)
{
	struct termios *p = &ios, rawios;

	if (raw) {
		if (tcgetattr(tty, &ios) == -1)
			return -1;
		rawios = ios;
		rawios.c_lflag &= ~(ICANON | ECHO);
		rawios.c_iflag &= ~(IXON);
		p = &rawios;
	}
	if (tcsetattr(tty, TCSANOW, p) == -1)
		return -1;
	return 0;
}

void
panic(const char *fmt, ...)
{
	static int panicking = 0;
	va_list ap;

	if (panicking)
		return;
	panicking = 1;

	if (tty && tty != -1) {
		tttidy();
		ttclose();
	}

	if (fmt) {
		va_start(ap, fmt);
		vfprintf(stderr, fmt, ap);
		va_end(ap);
		if (errno)
			fputs(": ", stderr);
	}
	if (errno)
		fprintf(stderr, strerror(errno));
	if (fmt || errno)
		fputc('\n', stderr);

	exit(1);
}

void
fini(void)
{
	if (tty && tty != -1) {
		tttidy();
		ttclose();
	}
}

int
main(void)
{
	errno = 0;

	if (ttopen("/dev/tty") == -1)
		err(1, "/dev/tty");
	atexit(fini);

	for (;;) {
		char c;
		display();
		if (read(tty, &c, 1) == -1)
			panic("read key");
		switch (c) {
		case CTRL('q'): goto end;
		}
	}
end:

	return 0;
}
