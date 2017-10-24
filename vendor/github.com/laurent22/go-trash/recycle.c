// +build windows

// Slightly modified version of recycle.c from Matt Ginzton - http://www.maddogsw.com/cmdutils/
// Mainly replaced BOOL type by int to make it easier to call the function from Go.
//
// Also added functions for conversion between Go []string and **char from John Barham:
// https://groups.google.com/forum/#!topic/golang-nuts/pQueMFdY0mk

#include <stdio.h>
#include <windows.h>
#include <stdlib.h>

char **makeCharArray(int size) {
	return calloc(sizeof(char*), size);
}

void setArrayString(char **a, char *s, int n) {
 	a[n] = s;
}

void freeCharArray(char **a, int size) {
	int i;
	for (i = 0; i < size; i++)
		free(a[i]);
	free(a);
}

int RecycleFiles (char** filenames, int nFiles, int bConfirmed)
{
	SHFILEOPSTRUCT opRecycle;
	char* pszFilesToRecycle;
	char* pszNext;
	int i, len;
	int success = 1;
	char szLongBuf[MAX_PATH];
	char* lastComponent;

	//fill filenames to delete
	len = 0;
	for (i = 0; i < nFiles; i++)
	{
		GetFullPathName (filenames[i], sizeof(szLongBuf), szLongBuf, &lastComponent);
		len += lstrlen (szLongBuf) + 1;
	}

	pszFilesToRecycle = malloc (len + 1);
	pszNext = pszFilesToRecycle;
	for (i = 0; i < nFiles; i++)
	{
		GetFullPathName (filenames[i], sizeof(szLongBuf), szLongBuf, &lastComponent);

		lstrcpy (pszNext, szLongBuf);
		pszNext += lstrlen (pszNext) + 1;		//advance past terminator
	}
	*pszNext = 0;		//double-terminate

	//fill fileop structure
	opRecycle.hwnd = NULL;
	opRecycle.wFunc = FO_DELETE;
	opRecycle.pFrom = pszFilesToRecycle;
	opRecycle.pTo = "\0\0";
	opRecycle.fFlags = FOF_ALLOWUNDO;
	if (bConfirmed)
		opRecycle.fFlags |= FOF_NOCONFIRMATION;
	opRecycle.lpszProgressTitle = "Recycling files...";

	if (0 != SHFileOperation (&opRecycle))
		success = 0;
	if (opRecycle.fAnyOperationsAborted)
		success = 0;

	free (pszFilesToRecycle);

	return success;
}