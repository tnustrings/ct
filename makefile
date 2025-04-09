# sketch

publish:

gh release create v<version> ./dist/ct-<version>-py3-none-any.whl

update an asset:

gh release delete-asset v<version> ct-<version>-py3-none-any.whl -y
gh release upload v<version> ./dist/ct-<version>-py3-none-any.whl
