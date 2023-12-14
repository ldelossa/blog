DROPBOX_DIR = ~/Dropbox/Obsidian/Obsidian/Blog

.PHONY:
link-obsidian:
	-mkdir $(DROPBOX_DIR)
	-unlink $(DROPBOX_DIR)/posts
	-unlink $(DROPBOX_DIR)/static
	ln -s $$(pwd)/content/posts $(DROPBOX_DIR)/posts
	ln -s $$(pwd)/static $(DROPBOX_DIR)/static
