"use client";
import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { Button } from "../ui/button";

const CodeBlock = ({ children, className }: { children: React.ReactNode[]; className: string }) => {
  const [copied, setCopied] = useState(false);
  const codeContent = children[0] || "";

  const handleCopy = async () => {
    if (typeof codeContent === "string") {
      await navigator.clipboard.writeText(codeContent);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="relative group">
      <pre className={className}>
        <code className={className}>{children}</code>
      </pre>
      <Button variant="link" onClick={handleCopy} className="absolute top-2 right-2  p-1.5 rounded-md opacity-0 group-hover:opacity-100 transition-opacity" aria-label="Copy to clipboard">
        {copied ? <Check size={16} /> : <Copy size={16} />}
      </Button>
    </div>
  );
};

export default CodeBlock;
