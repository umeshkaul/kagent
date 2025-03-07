import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { Model } from "@/lib/types";

interface ModelSelectionSectionProps {
  allModels: Model[];
  selectedModel: Model;
  setSelectedModel: (model: Model) => void;
  error?: string;
  isSubmitting: boolean;
}

export const ModelSelectionSection = ({ allModels, selectedModel, setSelectedModel, error, isSubmitting }: ModelSelectionSectionProps) => {
  return (
    <>
      <label className="text-sm mb-2 block">Model</label>
      <Select value={selectedModel.id} disabled={isSubmitting} onValueChange={(value) => setSelectedModel(allModels.find((m) => m.id === value) || allModels[0])}>
        <SelectTrigger className={`${error ? "border-red-500" : ""}`}>
          <SelectValue placeholder="Select a model" />
        </SelectTrigger>
        <SelectContent className="">
          {allModels.map((model) => (
            <SelectItem key={model.id} value={model.id} >
              {model.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {error && <p className="text-red-500 text-sm mt-1">{error}</p>}
    </>
  );
};
