import sys
# Just a bunch of prints
def func():
    msg = "Hello from 'func', Go" 
    print(msg)
    
print(sys.argv[1]) # Prints argument on index 1 ie argument 2
msg = "Hello Go"
print(msg)
func()
msg = "Bye Go"
print(msg)


