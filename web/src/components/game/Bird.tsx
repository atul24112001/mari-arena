import Matter, { Body } from "matter-js";
import React, { useEffect, useState } from "react";

export default function Bird({ size, body, color, fly }: Props) {
  const [width, height] = size;
  const x = body.position.x - width / 2;
  const y = body.position.y - height / 2;

  return (
    <div
      className={`${
        // fly
        //   ? "bg-[url('/assets/bird/redbird-upflap.png')]"
        //   :
        "bg-[url('/assets/bird/redbird-downflap.png')]"
      } bg-cover `}
      style={{
        position: "absolute",
        top: y,
        left: x,
        width,
        height,
        backgroundColor: "transparent",
        zIndex: 1000,
      }}
    />
  );
}

type Props = {
  size: [number, number];
  body: Body;
  color: string;
  fly?: boolean;
};
