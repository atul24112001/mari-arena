import Link from "next/link";
import React from "react";

export function Card({ href, title, description }: Props) {
  return (
    <Link href={href}>
      <div className="bg-[#2b924540] w-[290px] text-[#fff] rounded-md px-2 py-2 mb-2">
        <h1 className="font-bold text-2xl text-center  cursor-pointer">
          {title}
        </h1>
        {description && (
          <p className="text-center text-xs opacity-80">{description}</p>
        )}
      </div>
    </Link>
  );
}

type Props = {
  href: string;
  title: string;
  description?: string;
};
