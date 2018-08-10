#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>

int forkbomb(){
  unsigned char XD[20] = {'I', 't', '\'', 's', ' ', 'o', 'v', 'e', 'r', ' ', '9', '0', '0', '0', '!', '!', '\n', '\0', 1};
  unsigned char * lol = XD+18;
  
  while( *lol ){
    fork();
    printf("%s\n", XD + *lol);
    *lol = (*lol + 1 )%20;
  }

  return 0;
}

int main()
{
   printf("Testing sockets\n");
   system("wget 8.8.8.8");

   printf("Testing dir permissions\n");
   system("rm /rootowned");
   system("chown sandbox:sandbox /rootowned");
   system("rm /rootowned");

   printf("Running forkbomb\n");
   forkbomb();
   printf("Forkbomb survived.\n");
   return 0;
}