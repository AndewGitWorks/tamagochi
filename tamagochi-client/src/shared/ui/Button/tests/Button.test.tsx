import { render, screen } from '@testing-library/react';
import { Button } from '../Button';

describe('Button', () => {
  it('should render title', () => {
    render(<Button title="Hello Btn!" />);

    const btn = screen.getByRole('button', { name: 'Hello Btn!' });
    expect(btn).toBeInTheDocument();
  });
});
