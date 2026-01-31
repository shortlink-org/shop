import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import Label from '../label';

describe('Label', () => {
  it('should render title text', () => {
    render(<Label title="Test Product" amount={99.99} />);
    expect(screen.getByText('Test Product')).toBeInTheDocument();
  });

  it('should render with price', () => {
    render(<Label title="Product" amount={49.99} currencyCode="USD" />);
    expect(screen.getByText('USD')).toBeInTheDocument();
  });

  it('should render with default bottom position', () => {
    const { container } = render(<Label title="Product" amount={100} />);
    const labelDiv = container.firstChild as HTMLElement;
    expect(labelDiv).toHaveClass('bottom-0');
    expect(labelDiv).not.toHaveClass('lg:pb-[35%]');
  });

  it('should render with center position', () => {
    const { container } = render(<Label title="Product" amount={100} position="center" />);
    const labelDiv = container.firstChild as HTMLElement;
    expect(labelDiv).toHaveClass('lg:px-20');
    expect(labelDiv).toHaveClass('lg:pb-[35%]');
  });

  it('should render title in h3 element', () => {
    render(<Label title="Heading Test" amount={50} />);
    const heading = screen.getByRole('heading', { level: 3 });
    expect(heading).toHaveTextContent('Heading Test');
  });

  it('should apply line-clamp-2 to title for truncation', () => {
    render(<Label title="Very long product title that might overflow" amount={100} />);
    const heading = screen.getByRole('heading', { level: 3 });
    expect(heading).toHaveClass('line-clamp-2');
  });

  it('should render with different currencies', () => {
    render(<Label title="Product" amount={1500} currencyCode="RUB" />);
    expect(screen.getByText('RUB')).toBeInTheDocument();
  });

  it('should have backdrop blur styling', () => {
    const { container } = render(<Label title="Product" amount={100} />);
    const innerDiv = container.querySelector('.backdrop-blur-md');
    expect(innerDiv).toBeInTheDocument();
  });
});
