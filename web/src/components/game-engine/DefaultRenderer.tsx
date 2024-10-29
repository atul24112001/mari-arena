import React, { ReactNode } from "react";

type Renderer = (entities: any, window: Window) => ReactNode;

const DefaultRenderer: Renderer = (entities, window) => {
  if (!entities || !window) return null;

  return Object.keys(entities)
    .filter((key) => entities[key].renderer)
    .map((key) => {
      const entity = entities[key];
      if (typeof entity.renderer === "object") {
        return <entity.renderer.type key={key} window={window} {...entity} />;
      } else if (typeof entity.renderer === "function") {
        return <entity.renderer key={key} window={window} {...entity} />;
      }
      return null;
    });
};

export default DefaultRenderer;
