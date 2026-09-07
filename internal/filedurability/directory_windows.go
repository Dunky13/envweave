package filedurability

// SyncDirectory is a no-op on Windows: directory handles cannot be flushed
// (FlushFileBuffers fails with access denied), and NTFS journals the rename.
// compose/fsio_windows.go makes the same call.
func SyncDirectory(string) error { return nil }
