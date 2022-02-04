import * as React from "react";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import DialogTitle from "@mui/material/DialogTitle";
import Dialog from "@mui/material/Dialog";
import { colorschemes, ColorScheme } from "../colors/colorschemes";

type Props = {
  colorscheme: ColorScheme;
  open: boolean;
  onChange: (colorscheme: ColorScheme) => void;
  onClose: () => void;
};
function ColorChooser_({ colorscheme, open, onChange, onClose }: Props) {
  return (
    <Dialog onClose={onClose} open={open}>
      <DialogTitle>Color Scheme</DialogTitle>
      <List sx={{ pt: 0 }}>
        {Object.keys(colorschemes).map((cs) => (
          <ListItem
            button
            selected={colorscheme === cs}
            onClick={() => onChange(cs as ColorScheme)}
            key={cs}
          >
            <ListItemText primary={colorschemes[cs as ColorScheme]} />
          </ListItem>
        ))}
      </List>
    </Dialog>
  );
}

export default React.memo(ColorChooser_);
