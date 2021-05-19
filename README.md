# KRM Linter

A linter for the kubernetes resource model

Currently all it does parse a CRD yaml and generate CSV lines with info about
the CRD and a couple of simple "readability checks" (what is the apiVersion of
the CRD and does the CRD's spec have any OpenAPIV3Schema at all).
