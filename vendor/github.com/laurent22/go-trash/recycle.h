// +build windows

int RecycleFiles (char** filenames, int nFiles, int bConfirmed);
char** makeCharArray(int size);
void setArrayString(char **a, char *s, int n);
void freeCharArray(char **a, int size);