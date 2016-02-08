PKGS=$(wildcard pkg/*)
clean_PKGS=$(addprefix clean_,$(PKGS))

all: $(PKGS)
clean: $(clean_PKGS)

.PHONY: force
$(PKGS): force
	make -C $@

$(clean_PKGS): force
	make -C $(patsubst clean_%,%,$@) clean

examples:
	make -C examples
