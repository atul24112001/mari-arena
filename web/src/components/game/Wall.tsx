import { Body } from "matter-js";
import React from "react";

export default function Wall({ size, body, color }: Props) {
  const [width, height] = size;
  const x = body.position.x - width / 2;
  const y = body.position.y - height / 2;

  return (
    <div
      className={`bg-[url('/assets/base.png')] `}
      style={{
        position: "absolute",
        top: y,
        left: x,
        width,
        height,
        backgroundColor: color,
      }}
    />
  );
}

type Props = {
  size: [number, number];
  body: Body;
  color: string;
};
