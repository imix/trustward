# Device-property vocabulary stays model-local until reuse is proven

Component `properties:` that gate which EMB3D threats apply (e.g. reader `protocol`, controller `mgmt-interface`) are free-typed per model, not a shared schema or vocabulary in the renderer. The gating *mechanism* is generic, but the keys and allowed values are domain-specific; promoting them to a shared vocabulary is deferred until a second model needs the same gating. Generalising on a single example would be speculative.
