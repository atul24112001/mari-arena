import React from "react";
import ClientHomePage from "@/sections/home";
import { Card } from "@/components/helper";

export default function Home() {
  return (
    <div className="h-screen bg-[#e9f5ed] overflow-hidden bg-cover md:bg-contain  ">
      <div className="px-4 md:px-10 py-4 mb-4 gap-4 flex justify-between items-center">
        <h1 className="text-[#2b9245] text-sm md:text-lg   font-bold flex-1">
          Mari&nbsp;arena
        </h1>
        <ClientHomePage />
      </div>
      <div className="flex justify-center items-center">
        <div>
          <Card href="solo" title="Solo" description="Entry Free | Winner ₹0" />
          {/* <Card
          href={"1v1-inr"}
          title="1v1"
          description="Entry ₹10 | Winner ₹15"
        /> */}
          <Card
            href={"1v1-sol"}
            title="1v1"
            description="Entry 0.001 SOL | Winner 0.0015 SOL"
          />
          {/* <Card href="1v10-inr" title="1v10" description="Entry ₹15 | Winner ₹75" />
        <Card href="1v10-sol" title="1v10" description="Entry ₹15 | Winner ₹75" /> */}
        </div>
      </div>
    </div>
  );
}

{
  /* bg-[url('/assets/background-day.png')] */
}
