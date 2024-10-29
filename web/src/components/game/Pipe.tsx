import { Constants } from "@/lib/constants";
import { randomNumber } from "@/lib/functions";
import Matter from "matter-js";

import React from "react";

export const generatePipe = () => {
  const topPipe = randomNumber(100, Constants.MAX_HEIGHT / 2 - 100);
  const bottomPipe = Constants.MAX_HEIGHT - topPipe - Constants.PIPE_GAP_SIZE;

  let sizes = [topPipe, bottomPipe];
  if (Math.random() < 0.5) {
    sizes = sizes.reverse();
  }
  return sizes;
};

export default function Pipe({ size, body, color, top }: Props) {
  const [width, height] = size;
  const x = body.position.x - width / 2;
  const y = body.position.y - height / 2;

  return (
    <div
      className={`bg-[url('/assets/pipe-green.png')]  bg-cover ${
        top ? "rotate-180" : ""
      }`}
      style={{
        position: "absolute",
        top: y,
        left: x,
        width,
        height,
        backgroundColor: "transparent",
        display:
          x > Constants.MAX_WIDTH + Constants.PIPE_WIDTH ? "none" : "block",
      }}
    />
  );
}

type Props = {
  size: [number, number];
  body: Matter.Body;
  color: string;
  top?: boolean;
};
