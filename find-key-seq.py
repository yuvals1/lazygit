import sys
import tty
import termios


def getkey():
    old_settings = termios.tcgetattr(sys.stdin)
    try:
        tty.setraw(sys.stdin.fileno())
        ch = sys.stdin.read(1)
    finally:
        termios.tcsetattr(sys.stdin, termios.TCSADRAIN, old_settings)
    return ch


print("Press keys (Ctrl+C to exit):")
while True:
    key = getkey()
    print(f"Pressed: {ord(key)} ({repr(key)})")