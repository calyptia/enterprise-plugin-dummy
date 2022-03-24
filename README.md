# Calyptia Enterprise Dummy plugin

A Calyptia Enterprise plugin providing dummy input.

The canonical source of this repository including all copyright and licensing information is here: https://github.com/calyptia/enterprise-plugin-dummy
## Plugin architecture

The intention is that every plugin provides a tarball in its release per architecture.

This tarball will then be used in the Calyptia Enterprise Fluent Bit packaging process: the tarball will be extracted into a separate directory per plugin with the assumption that everything required is present.
