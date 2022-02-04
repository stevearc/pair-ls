import * as React from "react";
import Tabs from "./tabs";
import styled from "@emotion/styled";
import Menu from "./menu";

const Container = styled.div`
  display: flex;
  flex-direction: row;
  box-shadow: 0px 2px 5px rgba(1, 1, 1, 0.4);
  z-index: 10;
`;

export default function Header() {
  return (
    <Container>
      <Menu />
      <Tabs />
    </Container>
  );
}
