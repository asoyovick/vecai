# ==============================================================================
# VECAI Backend Makefile
# Delegates targets to the root Makefile so make commands work seamlessly
# when executed from within the backend directory.
# ==============================================================================

.PHONY: default
default:
	@$(MAKE) -C .. help

%:
	@$(MAKE) -C .. $@
