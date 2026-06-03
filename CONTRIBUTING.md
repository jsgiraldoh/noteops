# Contribuir a NoteOPs

## Branching

```
main        producción (protegida)
develop     integración continua
feature/*   nuevas funcionalidades
fix/*       correcciones
```

## Flujo de trabajo

```bash
git checkout develop
git checkout -b feature/nombre-descriptivo
# ... hacer cambios ...
git push origin feature/nombre-descriptivo
# Abrir PR hacia develop
```

## Convención de commits

```
feat: agregar reloj en tiempo real
fix: corregir cálculo de nota definitiva
docs: actualizar README
chore: actualizar dependencias
```

## Tests

```bash
make test
```

Los PRs deben pasar CI antes de ser mergeados.
